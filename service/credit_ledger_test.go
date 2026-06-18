package service

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
)

func TestAlignQuotaToDisplayPrecision(t *testing.T) {
	cs := operation_setting.GetCreditSetting()
	gs := operation_setting.GetGeneralSetting()

	// 保存并恢复全局状态
	origRound := cs.RoundToDisplayPrecision
	origType := gs.QuotaDisplayType
	origRate := gs.CustomCurrencyExchangeRate
	origPerUnit := common.QuotaPerUnit
	defer func() {
		cs.RoundToDisplayPrecision = origRound
		gs.QuotaDisplayType = origType
		gs.CustomCurrencyExchangeRate = origRate
		common.QuotaPerUnit = origPerUnit
	}()

	// 配置：QuotaPerUnit=500000，汇率=100（1元=100积分）。
	// 一分积分 = QuotaPerUnit/(100*100) = 50 quota。
	common.QuotaPerUnit = 500000
	gs.QuotaDisplayType = operation_setting.QuotaDisplayTypeCustom
	gs.CustomCurrencyExchangeRate = 100

	// 开关关闭：原样返回
	cs.RoundToDisplayPrecision = false
	if got := AlignQuotaToDisplayPrecision(123); got != 123 {
		t.Fatalf("disabled: want 123, got %d", got)
	}

	// 开关开启：向上对齐到 50 的整数倍
	cs.RoundToDisplayPrecision = true
	cases := []struct {
		in   int
		want int
	}{
		{0, 0},
		{1, 50},    // 0.0002 积分 → 向上取整到 0.01 积分 = 50 quota
		{50, 50},   // 恰好一分
		{51, 100},  // 略超一分 → 两分
		{100, 100}, // 恰好两分
	}
	for _, tc := range cases {
		if got := AlignQuotaToDisplayPrecision(tc.in); got != tc.want {
			t.Errorf("align(%d): want %d, got %d", tc.in, tc.want, got)
		}
	}

	// 非 CUSTOM 展示：原样返回
	gs.QuotaDisplayType = operation_setting.QuotaDisplayTypeUSD
	if got := AlignQuotaToDisplayPrecision(51); got != 51 {
		t.Fatalf("non-custom: want 51, got %d", got)
	}
}
