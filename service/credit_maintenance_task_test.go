/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
package service

import (
	"testing"

	"github.com/QuantumNous/new-api/setting/billing_setting"
)

// newEmptyPricingPatch 构造一个空映射的 patch，避免依赖全局 ratio_setting/DB。
func newEmptyPricingPatch() *pricingPatch {
	return &pricingPatch{
		price:           map[string]float64{},
		ratio:           map[string]float64{},
		cache:           map[string]float64{},
		createCache:     map[string]float64{},
		completion:      map[string]float64{},
		image:           map[string]float64{},
		audio:           map[string]float64{},
		audioCompletion: map[string]float64{},
		billingMode:     map[string]string{},
		billingExpr:     map[string]string{},
	}
}

func TestPricingPatchApplySnapshotPerToken(t *testing.T) {
	p := newEmptyPricingPatch()
	// 按 token：输入价 3 元/1M → ratio=1.5（前端已换算），补全倍率 2
	snap := `{"name":"gpt-x","billingMode":"per-token","ratio":"1.5","completionRatio":"2","cacheRatio":"0.5"}`
	if err := p.applySnapshot(snap); err != nil {
		t.Fatalf("applySnapshot error: %v", err)
	}
	if got := p.ratio["gpt-x"]; got != 1.5 {
		t.Errorf("ratio = %v, want 1.5", got)
	}
	if got := p.completion["gpt-x"]; got != 2 {
		t.Errorf("completion = %v, want 2", got)
	}
	if got := p.cache["gpt-x"]; got != 0.5 {
		t.Errorf("cache = %v, want 0.5", got)
	}
	if _, ok := p.price["gpt-x"]; ok {
		t.Errorf("per-token mode should not set price")
	}
}

func TestPricingPatchApplySnapshotPerRequest(t *testing.T) {
	p := newEmptyPricingPatch()
	// 按次：固定价 0.04 元/次。即便带了 ratio 也应只写 price。
	snap := `{"name":"mj","billingMode":"per-request","price":"0.04","ratio":"9"}`
	if err := p.applySnapshot(snap); err != nil {
		t.Fatalf("applySnapshot error: %v", err)
	}
	if got := p.price["mj"]; got != 0.04 {
		t.Errorf("price = %v, want 0.04", got)
	}
	if _, ok := p.ratio["mj"]; ok {
		t.Errorf("per-request mode should not set ratio")
	}
}

func TestPricingPatchApplySnapshotTieredExpr(t *testing.T) {
	p := newEmptyPricingPatch()
	snap := `{"name":"claude","billingMode":"tiered_expr","billingExpr":"tier(\"base\", p*1)","requestRuleExpr":"2","ratio":"1"}`
	if err := p.applySnapshot(snap); err != nil {
		t.Fatalf("applySnapshot error: %v", err)
	}
	if got := p.billingMode["claude"]; got != billing_setting.BillingModeTieredExpr {
		t.Errorf("billingMode = %q, want tiered_expr", got)
	}
	want := `(tier("base", p*1)) * 2`
	if got := p.billingExpr["claude"]; got != want {
		t.Errorf("billingExpr = %q, want %q", got, want)
	}
	// tiered_expr 保留 ratio 作 fallback
	if got := p.ratio["claude"]; got != 1 {
		t.Errorf("fallback ratio = %v, want 1", got)
	}
}

func TestPricingPatchApplySnapshotClearsStaleEntries(t *testing.T) {
	p := newEmptyPricingPatch()
	// 模型原先是按 token（有 ratio），改价为按次后，旧 ratio 应被清除
	p.ratio["m"] = 5
	p.completion["m"] = 3
	snap := `{"name":"m","billingMode":"per-request","price":"0.1"}`
	if err := p.applySnapshot(snap); err != nil {
		t.Fatalf("applySnapshot error: %v", err)
	}
	if _, ok := p.ratio["m"]; ok {
		t.Errorf("stale ratio should be cleared on mode switch")
	}
	if _, ok := p.completion["m"]; ok {
		t.Errorf("stale completion should be cleared on mode switch")
	}
	if got := p.price["m"]; got != 0.1 {
		t.Errorf("price = %v, want 0.1", got)
	}
}

func TestPricingPatchApplyLegacyRatio(t *testing.T) {
	p := newEmptyPricingPatch()
	p.applyLegacyRatio("old-model", 2.5, 1.5)
	if got := p.ratio["old-model"]; got != 2.5 {
		t.Errorf("legacy ratio = %v, want 2.5", got)
	}
	if got := p.completion["old-model"]; got != 1.5 {
		t.Errorf("legacy completion = %v, want 1.5", got)
	}
}

func TestCombineBillingExpr(t *testing.T) {
	cases := []struct {
		base, rules, want string
	}{
		{"tier(\"b\", p*1)", "2", `(tier("b", p*1)) * 2`},
		{"tier(\"b\", p*1)", "", `tier("b", p*1)`},
		{"", "2", ""},
		{"  base  ", "  3  ", "(base) * 3"},
	}
	for _, c := range cases {
		if got := combineBillingExpr(c.base, c.rules); got != c.want {
			t.Errorf("combineBillingExpr(%q,%q) = %q, want %q", c.base, c.rules, got, c.want)
		}
	}
}
