package controller

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"

	"github.com/gin-gonic/gin"
)

// creditSourceView 单个来源的展示数据。
type creditSourceView struct {
	Source        string  `json:"source"`
	Remaining     int     `json:"remaining"`      // quota
	RemainingUnit float64 `json:"remaining_unit"` // 积分
	NearestExpire int64   `json:"nearest_expire"`
}

// GetCreditBatches 返回当前用户积分构成（赠送/套餐/充值占比 + 有效期 + 即将过期提示）。
// 对应接口：GET /api/credit/batches
func GetCreditBatches(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		common.ApiErrorMsg(c, "未登录")
		return
	}

	summaries, err := model.GetCreditSourceSummaries(userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	rate := quotaToUnitRate()
	total := 0
	sources := make([]creditSourceView, 0, len(summaries))
	for _, s := range summaries {
		total += s.Remaining
		sources = append(sources, creditSourceView{
			Source:        s.Source,
			Remaining:     s.Remaining,
			RemainingUnit: quotaToUnit(s.Remaining, rate),
			NearestExpire: s.NearestExpire,
		})
	}

	data := gin.H{
		"sources":       sources,
		"total":         total,
		"total_unit":    quotaToUnit(total, rate),
		"currency_unit": operation_setting.GetCurrencySymbol(),
	}

	// 即将过期提示
	if nearest, err := model.NearestExpiringBatch(userId); err == nil && nearest != nil {
		data["expiring_soon"] = gin.H{
			"remaining":      nearest.Remaining,
			"remaining_unit": quotaToUnit(nearest.Remaining, rate),
			"expire_at":      nearest.ExpireAt,
			"source":         nearest.Source,
		}
	}

	common.ApiSuccess(c, data)
}

// quotaToUnitRate 返回 quota→积分 的换算系数（积分 = quota × rate）。
func quotaToUnitRate() float64 {
	if common.QuotaPerUnit <= 0 {
		return 0
	}
	return operation_setting.GetGeneralSetting().CustomCurrencyExchangeRate / common.QuotaPerUnit
}

func quotaToUnit(quota int, rate float64) float64 {
	return float64(quota) * rate
}
