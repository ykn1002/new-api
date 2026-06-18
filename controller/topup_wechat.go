package controller

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"

	"github.com/gin-gonic/gin"
)

const wechatPayUnifiedOrderURL = "https://api.mch.weixin.qq.com/v3/pay/transactions/jsapi"

// isWeChatPayEnabled 微信支付是否可用（合规确认 + 配置完整）。
func isWeChatPayEnabled() bool {
	return operation_setting.IsPaymentComplianceConfirmed() &&
		operation_setting.GetWeChatPaySetting().IsConfigured()
}

type weChatPayRequest struct {
	Amount int64  `json:"amount"` // 充值金额（元）
	TierId string `json:"tier_id"`
	OpenId string `json:"openid"` // 小程序用户 openid（JSAPI 必填）
}

// RequestWeChatPay 微信支付（JSAPI/小程序）统一下单，返回前端调起支付所需签名参数。
// POST /api/topup/wechat
func RequestWeChatPay(c *gin.Context) {
	if !isWeChatPayEnabled() {
		common.ApiErrorMsg(c, "微信支付未开启")
		return
	}
	var req weChatPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	if req.OpenId == "" {
		common.ApiErrorMsg(c, "缺少 openid")
		return
	}

	cfg := operation_setting.GetWeChatPaySetting()
	userId := c.GetInt("id")

	// 解析金额：命中档位用档位金额，否则用自由金额
	money := float64(req.Amount)
	if req.TierId != "" {
		if tier := operation_setting.FindRechargeTier(req.TierId); tier != nil {
			money = tier.Amount
		}
	}
	if money < float64(operation_setting.MinTopUp) {
		common.ApiErrorMsg(c, fmt.Sprintf("充值金额不能小于 %d 元", operation_setting.MinTopUp))
		return
	}

	// 充值额度 quota = 金额 × QuotaPerUnit（与「人民币≡美元」口径一致）
	quota := int64(money * common.QuotaPerUnit)
	tradeNo := fmt.Sprintf("wxpay_%d_%d", userId, time.Now().UnixNano())

	topUp := &model.TopUp{
		UserId:          userId,
		Amount:          quota,
		Money:           money,
		TradeNo:         tradeNo,
		PaymentMethod:   model.PaymentMethodWeChat,
		PaymentProvider: model.PaymentProviderWeChat,
		CreateTime:      time.Now().Unix(),
		Status:          common.TopUpStatusPending,
	}
	if err := topUp.Insert(); err != nil {
		common.ApiError(c, err)
		return
	}

	prepayId, err := wechatPayUnifiedOrder(cfg, tradeNo, req.OpenId, money, "积分充值")
	if err != nil {
		logger.LogError(c.Request.Context(), "微信支付下单失败: "+err.Error())
		common.ApiErrorMsg(c, "拉起支付失败")
		return
	}

	payParams, err := buildMiniProgramPaySign(cfg, prepayId)
	if err != nil {
		logger.LogError(c.Request.Context(), "微信支付签名失败: "+err.Error())
		common.ApiErrorMsg(c, "支付签名失败")
		return
	}
	payParams["order_id"] = tradeNo
	common.ApiSuccess(c, payParams)
}

// wechatPayUnifiedOrder 调用微信「JSAPI 统一下单」，返回 prepay_id。金额单位转为分。
func wechatPayUnifiedOrder(cfg *operation_setting.WeChatPaySetting, tradeNo, openId string, money float64, desc string) (string, error) {
	totalFen := int64(money * 100)
	body := map[string]interface{}{
		"appid":        cfg.AppId,
		"mchid":        cfg.MchId,
		"description":  desc,
		"out_trade_no": tradeNo,
		"notify_url":   cfg.NotifyUrl,
		"amount": map[string]interface{}{
			"total":    totalFen,
			"currency": "CNY",
		},
		"payer": map[string]interface{}{
			"openid": openId,
		},
	}
	bodyBytes, err := common.Marshal(body)
	if err != nil {
		return "", err
	}

	authHeader, err := wechatPayAuthorization(cfg, http.MethodPost, "/v3/pay/transactions/jsapi", string(bodyBytes))
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequest(http.MethodPost, wechatPayUnifiedOrderURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", authHeader)

	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("微信下单返回 %d: %s", resp.StatusCode, string(respBody))
	}
	var res struct {
		PrepayId string `json:"prepay_id"`
	}
	if err := common.Unmarshal(respBody, &res); err != nil {
		return "", err
	}
	if res.PrepayId == "" {
		return "", errors.New("微信未返回 prepay_id")
	}
	return res.PrepayId, nil
}

// wechatPayAuthorization 构造 APIv3 请求签名 Authorization 头。
func wechatPayAuthorization(cfg *operation_setting.WeChatPaySetting, method, urlPath, body string) (string, error) {
	nonce := common.GetRandomString(32)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	message := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n", method, urlPath, timestamp, nonce, body)

	signature, err := rsaSignWithPrivateKey(cfg.MchPrivateKey, message)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		`WECHATPAY2-SHA256-RSA2048 mchid="%s",nonce_str="%s",timestamp="%s",serial_no="%s",signature="%s"`,
		cfg.MchId, nonce, timestamp, cfg.MchCertSerialNo, signature,
	), nil
}

// buildMiniProgramPaySign 生成小程序 wx.requestPayment 所需的签名参数。
func buildMiniProgramPaySign(cfg *operation_setting.WeChatPaySetting, prepayId string) (map[string]interface{}, error) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := common.GetRandomString(32)
	pkg := "prepay_id=" + prepayId
	message := fmt.Sprintf("%s\n%s\n%s\n%s\n", cfg.AppId, timestamp, nonce, pkg)
	signature, err := rsaSignWithPrivateKey(cfg.MchPrivateKey, message)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"timeStamp": timestamp,
		"nonceStr":  nonce,
		"package":   pkg,
		"signType":  "RSA",
		"paySign":   signature,
	}, nil
}

// rsaSignWithPrivateKey 用商户私钥（PKCS#8 PEM）对 message 做 SHA256-RSA 签名，返回 base64。
func rsaSignWithPrivateKey(privateKeyPEM, message string) (string, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return "", errors.New("无效的商户私钥 PEM")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return "", errors.New("商户私钥非 RSA")
	}
	h := sha256.Sum256([]byte(message))
	sig, err := rsa.SignPKCS1v15(rand.Reader, rsaKey, crypto.SHA256, h[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}
