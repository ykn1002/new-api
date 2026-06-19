package operation_setting

import "github.com/QuantumNous/new-api/setting/config"

// CreditSetting 汇总 Token Plan（积分产品化）相关的运营配置项。
// 统一经全局配置管理器持久化到 Option 表（键前缀 credit_setting.*），
// 仅超管可在「全局策略 / 计费设置」页读写（路由层 RootAuth 已收口）。
type CreditSetting struct {
	// RoundToDisplayPrecision 开启后，单次模型调用的实扣额度会按「积分 2 位小数向上取整」对齐。
	// 仅在 QuotaDisplayType=CUSTOM（积分）时生效；关闭时维持原有整数 quota 结算口径。
	RoundToDisplayPrecision bool `json:"round_to_display_precision"`

	// GiftDefaultValidityDays 赠送积分（gift 批次）的默认有效期天数；0=永久。
	GiftDefaultValidityDays int `json:"gift_default_validity_days"`
	// TopupDefaultValidityDays 充值积分（topup 批次）在档位未指定有效期时的默认天数；0=永久。
	TopupDefaultValidityDays int `json:"topup_default_validity_days"`

	// LowBalanceWarnPercent 低余额预警百分比阈值（剩余/基准 < 该百分比触发），取值 0~100；0=关闭百分比预警。
	LowBalanceWarnPercent float64 `json:"low_balance_warn_percent"`
	// NotifyRootOnLowBalance 用户低余额时是否同时通知运营（复用 NotifyRootUser）。
	NotifyRootOnLowBalance bool `json:"notify_root_on_low_balance"`

	// ExpireScanIntervalMinutes 到期清零 / 价格定时生效扫描周期（分钟），最小 1。
	ExpireScanIntervalMinutes int `json:"expire_scan_interval_minutes"`

	// DefaultModelEnabled 开启「全局默认模型」后，忽略客户端传入的 model，强制路由到 DefaultModel。
	DefaultModelEnabled bool `json:"default_model_enabled"`
	// DefaultModel 全局默认模型名称。
	DefaultModel string `json:"default_model"`
	// EnforceModelStatus 开启后，被下线（Model.Status != 1）的模型不可调用（distributor 门禁）。
	EnforceModelStatus bool `json:"enforce_model_status"`
}

var creditSetting = CreditSetting{
	RoundToDisplayPrecision:   true,
	GiftDefaultValidityDays:   30,
	TopupDefaultValidityDays:  0,
	LowBalanceWarnPercent:     20,
	NotifyRootOnLowBalance:    false,
	ExpireScanIntervalMinutes: 5,
	DefaultModelEnabled:       false,
	DefaultModel:              "",
	EnforceModelStatus:        false,
}

func init() {
	config.GlobalConfig.Register("credit_setting", &creditSetting)
}

func GetCreditSetting() *CreditSetting {
	return &creditSetting
}

// GetExpireScanInterval 返回到期扫描周期（分钟），下限 1。
func (s *CreditSetting) GetExpireScanIntervalMinutes() int {
	if s.ExpireScanIntervalMinutes < 1 {
		return 1
	}
	return s.ExpireScanIntervalMinutes
}
