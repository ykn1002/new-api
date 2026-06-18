package operation_setting

import "github.com/QuantumNous/new-api/setting/config"

// WeChatMiniProgramSetting 微信小程序原生登录配置（缺口 A-6）。
// 直连微信 jscode2session（appid+secret+code → openid/unionid）。敏感值仅超管可配。
type WeChatMiniProgramSetting struct {
	Enabled bool   `json:"enabled"`
	AppId   string `json:"app_id"`
	Secret  string `json:"secret"`
}

var wechatMiniProgramSetting = WeChatMiniProgramSetting{}

func init() {
	config.GlobalConfig.Register("wechat_miniprogram_setting", &wechatMiniProgramSetting)
}

func GetWeChatMiniProgramSetting() *WeChatMiniProgramSetting {
	return &wechatMiniProgramSetting
}

func (s *WeChatMiniProgramSetting) IsConfigured() bool {
	return s.Enabled && s.AppId != "" && s.Secret != ""
}
