package operation_setting

import "github.com/QuantumNous/new-api/setting/config"

// WeChatPaySetting 微信支付（APIv3 / JSAPI 小程序支付）配置（缺口 A-5）。
// 敏感值（APIv3Key/商户私钥）仅超管可配，前端不回显原文。
type WeChatPaySetting struct {
	Enabled bool `json:"enabled"`
	// AppId 小程序 appid（与登录小程序一致，用于 JSAPI 下单的 payer/sub_appid）
	AppId string `json:"app_id"`
	// MchId 商户号
	MchId string `json:"mch_id"`
	// APIv3Key APIv3 密钥（用于回调解密）
	APIv3Key string `json:"api_v3_key"`
	// MchCertSerialNo 商户证书序列号（请求签名头 serial_no）
	MchCertSerialNo string `json:"mch_cert_serial_no"`
	// MchPrivateKey 商户 API 私钥（PEM，PKCS#8），用于请求签名
	MchPrivateKey string `json:"mch_private_key"`
	// PlatformPublicKey 微信支付平台证书公钥（PEM），用于回调验签
	PlatformPublicKey string `json:"platform_public_key"`
	// PlatformCertSerialNo 平台证书/公钥序列号，与回调头 Wechatpay-Serial 比对
	PlatformCertSerialNo string `json:"platform_cert_serial_no"`
	// NotifyUrl 支付结果通知回调 URL
	NotifyUrl string `json:"notify_url"`
}

var wechatPaySetting = WeChatPaySetting{}

func init() {
	config.GlobalConfig.Register("wechatpay_setting", &wechatPaySetting)
}

func GetWeChatPaySetting() *WeChatPaySetting {
	return &wechatPaySetting
}

func (s *WeChatPaySetting) IsConfigured() bool {
	return s.Enabled && s.AppId != "" && s.MchId != "" &&
		s.APIv3Key != "" && s.MchCertSerialNo != "" && s.MchPrivateKey != ""
}
