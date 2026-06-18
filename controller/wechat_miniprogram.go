package controller

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// WeChatMiniProgramProviderId 微信小程序在 UserOAuthBinding 中的稳定 provider_id。
// 用一个固定的大常量，避免与自增的自定义 OAuth provider id 冲突。
const WeChatMiniProgramProviderId = 1000001

type miniProgramLoginRequest struct {
	Code string `json:"code"`
}

type jscode2sessionResponse struct {
	OpenId     string `json:"openid"`
	UnionId    string `json:"unionid"`
	SessionKey string `json:"session_key"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// jscode2session 直连微信接口用 code 换 openid/unionid。
func jscode2session(appId, secret, code string) (*jscode2sessionResponse, error) {
	if code == "" {
		return nil, errors.New("无效的 code")
	}
	endpoint := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		url.QueryEscape(appId), url.QueryEscape(secret), url.QueryEscape(code),
	)
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res jscode2sessionResponse
	if err := common.DecodeJson(resp.Body, &res); err != nil {
		return nil, err
	}
	if res.ErrCode != 0 {
		return nil, fmt.Errorf("微信登录失败：%d %s", res.ErrCode, res.ErrMsg)
	}
	if res.OpenId == "" {
		return nil, errors.New("微信未返回 openid")
	}
	return &res, nil
}

// WeChatMiniProgramLogin 小程序原生登录 + 账号一对一映射 + 自动开户。
// POST /api/wechat/miniprogram/login { code }
func WeChatMiniProgramLogin(c *gin.Context) {
	cfg := operation_setting.GetWeChatMiniProgramSetting()
	if !cfg.IsConfigured() {
		common.ApiErrorMsg(c, "管理员未配置微信小程序登录")
		return
	}

	var req miniProgramLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "无效的请求")
		return
	}

	session, err := jscode2session(cfg.AppId, cfg.Secret, req.Code)
	if err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}

	// unionid 优先，回退 openid，作为稳定外部账号标识
	providerUserId := session.UnionId
	if providerUserId == "" {
		providerUserId = session.OpenId
	}

	// 1) 命中绑定 → 直接登录
	user, err := model.GetUserByOAuthBinding(WeChatMiniProgramProviderId, providerUserId)
	if err == nil && user != nil {
		if user.Status != common.UserStatusEnabled {
			common.ApiErrorMsg(c, "用户已被封禁")
			return
		}
		setupLogin(user, c)
		return
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		common.ApiError(c, err)
		return
	}

	// 2) 未命中 → 注册开关校验 → 事务内建号 + 绑定（唯一索引兜并发）
	if !common.RegisterEnabled {
		common.ApiErrorMsg(c, "管理员关闭了新用户注册")
		return
	}

	newUser, apiErr := createMiniProgramUser(providerUserId)
	if apiErr != nil {
		common.ApiErrorMsg(c, apiErr.Error())
		return
	}
	setupLogin(newUser, c)
}

// createMiniProgramUser 在事务内创建用户并绑定小程序账号；并发冲突时重查已建用户。
func createMiniProgramUser(providerUserId string) (*model.User, error) {
	user := &model.User{
		Username:    "wxmp_" + strconv.Itoa(model.GetMaxUserId()+1),
		DisplayName: "微信用户",
		Role:        common.RoleCommonUser,
		Status:      common.UserStatusEnabled,
	}

	txErr := model.DB.Transaction(func(tx *gorm.DB) error {
		if err := user.InsertWithTx(tx, 0); err != nil {
			return err
		}
		binding := &model.UserOAuthBinding{
			UserId:         user.Id,
			ProviderId:     WeChatMiniProgramProviderId,
			ProviderUserId: providerUserId,
		}
		return model.CreateUserOAuthBindingWithTx(tx, binding)
	})

	if txErr != nil {
		// 并发：绑定已被其它请求建立 → 重查取已存在用户
		if existing, err := model.GetUserByOAuthBinding(WeChatMiniProgramProviderId, providerUserId); err == nil && existing != nil {
			return existing, nil
		}
		return nil, txErr
	}

	// 新用户初始赠送额度（QuotaForNewUser）登记为 gift 批次，维护不变量 I
	if user.Quota > 0 {
		if err := model.RegisterTopupCreditBatch(user.Id, user.Quota, "new_user_bonus", 0); err != nil {
			common.SysLog("failed to register new miniprogram user bonus batch: " + err.Error())
		}
	}
	return user, nil
}
