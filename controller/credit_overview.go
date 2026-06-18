package controller

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"

	"github.com/gin-gonic/gin"
)

// GetCreditOverview 流量池总览聚合（缺口 B-13）。
// query: start_timestamp / end_timestamp
func GetCreditOverview(c *gin.Context) {
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	if endTimestamp == 0 {
		endTimestamp = common.GetTimestamp()
	}

	ov, err := model.GetCreditOverview(startTimestamp, endTimestamp)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	rate := quotaToUnitRate()
	// 成本估算 = 消耗 Token 按售价折算的总金额（quota → 金额，元）
	var costEstimate float64
	if common.QuotaPerUnit > 0 {
		costEstimate = float64(ov.TotalQuota) / common.QuotaPerUnit * operation_setting.USDExchangeRate
	}
	modelCredits := make(map[string]float64, len(ov.ModelQuota))
	for m, q := range ov.ModelQuota {
		modelCredits[m] = quotaToUnit(int(q), rate)
	}

	common.ApiSuccess(c, gin.H{
		"total_tokens":      ov.TotalTokens,
		"total_quota":       ov.TotalQuota,
		"total_credits":     quotaToUnit(int(ov.TotalQuota), rate),
		"cost_estimate":     costEstimate,
		"total_recharge":    ov.TotalRecharge,
		"user_count":        ov.UserCount,
		"active_user_count": ov.ActiveUserCount,
		"model_tokens":      ov.ModelTokens,
		"model_credits":     modelCredits,
		"currency_unit":     operation_setting.GetCurrencySymbol(),
	})
}

// GetUserCreditStats 用户列表按状态分类计数（缺口 B-16）。
func GetUserCreditStats(c *gin.Context) {
	warnPercent := operation_setting.GetCreditSetting().LowBalanceWarnPercent
	stats, err := model.GetUserCreditStats(warnPercent)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, stats)
}

// ExportCreditData 导出 CSV（缺口：数据导出）。
// query: type=model|user（默认 model），start_timestamp/end_timestamp。
func ExportCreditData(c *gin.Context) {
	exportType := c.DefaultQuery("type", "model")
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	if endTimestamp == 0 {
		endTimestamp = common.GetTimestamp()
	}

	rate := quotaToUnitRate()
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"credit_%s_export.csv\"", exportType))
	// UTF-8 BOM，便于 Excel 正确识别中文
	_, _ = c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(c.Writer)
	defer w.Flush()

	switch exportType {
	case "user":
		dates, err := model.GetQuotaDataGroupByUser(startTimestamp, endTimestamp)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		_ = w.Write([]string{"用户名", "时间", "调用次数", "消耗Token", "消耗积分"})
		for _, d := range dates {
			_ = w.Write([]string{
				d.Username,
				strconv.FormatInt(d.CreatedAt, 10),
				strconv.Itoa(d.Count),
				strconv.Itoa(d.TokenUsed),
				strconv.FormatFloat(quotaToUnit(d.Quota, rate), 'f', 2, 64),
			})
		}
	default: // model
		dates, err := model.GetAllQuotaDates(startTimestamp, endTimestamp, "")
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		_ = w.Write([]string{"模型", "时间", "调用次数", "消耗Token", "消耗积分"})
		for _, d := range dates {
			_ = w.Write([]string{
				d.ModelName,
				strconv.FormatInt(d.CreatedAt, 10),
				strconv.Itoa(d.Count),
				strconv.Itoa(d.TokenUsed),
				strconv.FormatFloat(quotaToUnit(d.Quota, rate), 'f', 2, 64),
			})
		}
	}
}
