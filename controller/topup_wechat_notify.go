package controller

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"

	"github.com/gin-gonic/gin"
)

// wechatPayNotifyBody 微信支付结果通知请求体（APIv3）。
type wechatPayNotifyBody struct {
	Id           string `json:"id"`
	EventType    string `json:"event_type"`
	ResourceType string `json:"resource_type"`
	Resource     struct {
		Algorithm      string `json:"algorithm"`
		Ciphertext     string `json:"ciphertext"`
		AssociatedData string `json:"associated_data"`
		Nonce          string `json:"nonce"`
		OriginalType   string `json:"original_type"`
	} `json:"resource"`
}

// wechatPayDecryptedResource 解密后的交易资源。
type wechatPayDecryptedResource struct {
	OutTradeNo    string `json:"out_trade_no"`
	TransactionId string `json:"transaction_id"`
	TradeState    string `json:"trade_state"`
	Amount        struct {
		Total    int64  `json:"total"`
		Currency string `json:"currency"`
	} `json:"amount"`
}

// WeChatPayNotify 微信支付结果通知回调（验签 → 解密 → 到账）。
// POST /api/topup/wechat/notify
func WeChatPayNotify(c *gin.Context) {
	if !isWeChatPayEnabled() {
		c.JSON(http.StatusForbidden, gin.H{"code": "FAIL", "message": "微信支付未开启"})
		return
	}
	cfg := operation_setting.GetWeChatPaySetting()

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "读取请求失败"})
		return
	}

	// 1) 验签（平台证书公钥 + 应答签名头）
	if err := verifyWeChatPaySignature(c, cfg, bodyBytes); err != nil {
		logger.LogError(c.Request.Context(), "微信支付回调验签失败: "+err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{"code": "FAIL", "message": "验签失败"})
		return
	}

	// 2) 解析 + 解密资源
	var notify wechatPayNotifyBody
	if err := common.Unmarshal(bodyBytes, &notify); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "解析失败"})
		return
	}
	plaintext, err := decryptWeChatPayResource(cfg.APIv3Key, notify.Resource.AssociatedData, notify.Resource.Nonce, notify.Resource.Ciphertext)
	if err != nil {
		logger.LogError(c.Request.Context(), "微信支付回调解密失败: "+err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "解密失败"})
		return
	}
	var resource wechatPayDecryptedResource
	if err := common.Unmarshal(plaintext, &resource); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "资源解析失败"})
		return
	}

	if resource.TradeState != "SUCCESS" {
		// 非成功状态直接 200 应答，避免微信重试
		c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "OK"})
		return
	}

	// 3) 幂等到账
	tradeNo := resource.OutTradeNo
	LockOrder(tradeNo)
	defer UnlockOrder(tradeNo)

	topUp := model.GetTopUpByTradeNo(tradeNo)
	if topUp == nil {
		c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "OK"}) // 订单不存在也应答成功，避免无限重试
		return
	}
	if topUp.Status != common.TopUpStatusPending {
		c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "OK"}) // 已处理
		return
	}
	// 金额二次校验（分）
	if resource.Amount.Total != int64(topUp.Money*100) {
		logger.LogError(c.Request.Context(), fmt.Sprintf("微信支付金额不匹配 trade_no=%s notify=%d order=%.2f", tradeNo, resource.Amount.Total, topUp.Money))
		c.JSON(http.StatusOK, gin.H{"code": "FAIL", "message": "金额不匹配"})
		return
	}

	if err := model.RechargeWeChat(tradeNo, c.ClientIP()); err != nil {
		logger.LogError(c.Request.Context(), "微信支付到账失败: "+err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "到账失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "OK"})
}

// verifyWeChatPaySignature 用平台证书公钥验证回调签名。
func verifyWeChatPaySignature(c *gin.Context, cfg *operation_setting.WeChatPaySetting, body []byte) error {
	timestamp := c.GetHeader("Wechatpay-Timestamp")
	nonce := c.GetHeader("Wechatpay-Nonce")
	signature := c.GetHeader("Wechatpay-Signature")
	serial := c.GetHeader("Wechatpay-Serial")
	if timestamp == "" || nonce == "" || signature == "" {
		return errors.New("缺少验签头")
	}
	if cfg.PlatformPublicKey == "" {
		return errors.New("未配置平台证书公钥")
	}
	if cfg.PlatformCertSerialNo != "" && serial != "" && cfg.PlatformCertSerialNo != serial {
		return fmt.Errorf("平台证书序列号不匹配: %s", serial)
	}

	message := fmt.Sprintf("%s\n%s\n%s\n", timestamp, nonce, string(body))
	sigBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return err
	}
	pubKey, err := parseWeChatPayPublicKey(cfg.PlatformPublicKey)
	if err != nil {
		return err
	}
	h := sha256.Sum256([]byte(message))
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, h[:], sigBytes)
}

// parseWeChatPayPublicKey 从 PEM 解析平台公钥（支持证书或公钥 PEM）。
func parseWeChatPayPublicKey(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("无效的平台证书/公钥 PEM")
	}
	// 优先按证书解析
	if cert, err := x509.ParseCertificate(block.Bytes); err == nil {
		if pub, ok := cert.PublicKey.(*rsa.PublicKey); ok {
			return pub, nil
		}
		return nil, errors.New("证书公钥非 RSA")
	}
	// 回退按 PKIX 公钥解析
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("公钥非 RSA")
	}
	return rsaPub, nil
}

// decryptWeChatPayResource 用 APIv3 key 做 AES-256-GCM 解密回调资源。
func decryptWeChatPayResource(apiV3Key, associatedData, nonce, ciphertext string) ([]byte, error) {
	if len(apiV3Key) != 32 {
		return nil, errors.New("APIv3 密钥长度必须为 32 字节")
	}
	cipherBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher([]byte(apiV3Key))
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(nonce) != gcm.NonceSize() {
		return nil, errors.New("nonce 长度不正确")
	}
	return gcm.Open(nil, []byte(nonce), cipherBytes, []byte(associatedData))
}
