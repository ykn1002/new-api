package operation_setting

import (
	"sort"

	"github.com/QuantumNous/new-api/setting/config"
)

// RechargeTier 结构化充值档位（缺口 A-3）：名称 + 金额 + 基础积分 + 赠送积分 + 有效期 + 上下架。
type RechargeTier struct {
	Id           string  `json:"id"`            // 稳定标识
	Name         string  `json:"name"`          // 档位名称：标准档/进阶档/旗舰档
	Amount       float64 `json:"amount"`        // 充值金额（元）
	BaseCredits  int64   `json:"base_credits"`  // 基础积分（= Amount × 汇率，可后台覆写）
	GiftCredits  int64   `json:"gift_credits"`  // 赠送积分（可选）
	ValidityDays int     `json:"validity_days"` // 基础积分有效期天数；0=永久
	Status       int     `json:"status"`        // 1=上架 0=下架
	SortOrder    int     `json:"sort_order"`
}

type PaymentSetting struct {
	AmountOptions  []int           `json:"amount_options"`
	AmountDiscount map[int]float64 `json:"amount_discount"` // 充值金额对应的折扣，例如 100 元 0.9 表示 100 元充值享受 9 折优惠

	// RechargeTiers 结构化充值档位列表（与 AmountOptions/AmountDiscount 并存）。
	RechargeTiers []RechargeTier `json:"recharge_tiers"`

	ComplianceConfirmed    bool   `json:"compliance_confirmed"`
	ComplianceTermsVersion string `json:"compliance_terms_version"`
	ComplianceConfirmedAt  int64  `json:"compliance_confirmed_at"`
	ComplianceConfirmedBy  int    `json:"compliance_confirmed_by"`
	ComplianceConfirmedIP  string `json:"compliance_confirmed_ip"`
}

const CurrentComplianceTermsVersion = "v1"

// 默认配置
var paymentSetting = PaymentSetting{
	AmountOptions:  []int{10, 20, 50, 100, 200, 500},
	AmountDiscount: map[int]float64{},
	RechargeTiers:  []RechargeTier{},
}

func init() {
	// 注册到全局配置管理器
	config.GlobalConfig.Register("payment_setting", &paymentSetting)
}

func GetPaymentSetting() *PaymentSetting {
	return &paymentSetting
}

func IsPaymentComplianceConfirmed() bool {
	return paymentSetting.ComplianceConfirmed &&
		paymentSetting.ComplianceTermsVersion == CurrentComplianceTermsVersion
}

// GetActiveRechargeTiers 返回已上架的充值档位，按 SortOrder 升序。
func GetActiveRechargeTiers() []RechargeTier {
	tiers := make([]RechargeTier, 0, len(paymentSetting.RechargeTiers))
	for _, t := range paymentSetting.RechargeTiers {
		if t.Status == 1 {
			tiers = append(tiers, t)
		}
	}
	sort.SliceStable(tiers, func(i, j int) bool {
		return tiers[i].SortOrder < tiers[j].SortOrder
	})
	return tiers
}

// FindRechargeTier 按 id 查找充值档位（含下架），未命中返回 nil。
func FindRechargeTier(id string) *RechargeTier {
	for i := range paymentSetting.RechargeTiers {
		if paymentSetting.RechargeTiers[i].Id == id {
			t := paymentSetting.RechargeTiers[i]
			return &t
		}
	}
	return nil
}

