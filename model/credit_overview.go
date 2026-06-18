package model

import (
	"github.com/QuantumNous/new-api/common"
)

// CreditOverview 流量池总览聚合（缺口 B-13）。
type CreditOverview struct {
	TotalTokens     int64            `json:"total_tokens"`     // 消耗 Token 总数
	TotalQuota      int64            `json:"total_quota"`      // 消耗 quota 总数（成本估算/积分换算的底）
	TotalRecharge   float64          `json:"total_recharge"`   // 累计充值金额（元）
	UserCount       int64            `json:"user_count"`       // 用户总数
	ActiveUserCount int64            `json:"active_user_count"`// 周期内有消费的用户数
	ModelTokens     map[string]int64 `json:"model_tokens"`     // 各模型消耗 token（占比用）
	ModelQuota      map[string]int64 `json:"model_quota"`      // 各模型消耗 quota
}

// GetCreditOverview 聚合周期内的消耗/充值/用户数据。
func GetCreditOverview(startTime, endTime int64) (*CreditOverview, error) {
	ov := &CreditOverview{
		ModelTokens: map[string]int64{},
		ModelQuota:  map[string]int64{},
	}

	// 1) 按模型聚合消耗（token / quota）
	dates, err := GetAllQuotaDates(startTime, endTime, "")
	if err != nil {
		return nil, err
	}
	for _, d := range dates {
		ov.TotalTokens += int64(d.TokenUsed)
		ov.TotalQuota += int64(d.Quota)
		ov.ModelTokens[d.ModelName] += int64(d.TokenUsed)
		ov.ModelQuota[d.ModelName] += int64(d.Quota)
	}

	// 2) 累计充值金额（成功订单）
	var recharge float64
	if err := DB.Model(&TopUp{}).
		Where("status = ? AND create_time >= ? AND create_time <= ?", common.TopUpStatusSuccess, startTime, endTime).
		Select("COALESCE(SUM(money),0)").Scan(&recharge).Error; err != nil {
		return nil, err
	}
	ov.TotalRecharge = recharge

	// 3) 用户总数
	if err := DB.Model(&User{}).Count(&ov.UserCount).Error; err != nil {
		return nil, err
	}

	// 4) 活跃用户数（周期内有消费的 distinct user）
	var activeCount int64
	if err := DB.Model(&QuotaData{}).
		Where("created_at >= ? AND created_at <= ?", startTime, endTime).
		Distinct("user_id").Count(&activeCount).Error; err != nil {
		return nil, err
	}
	ov.ActiveUserCount = activeCount

	return ov, nil
}

// UserCreditStats 用户列表按状态分类计数（缺口 B-16）。
type UserCreditStats struct {
	Total      int64 `json:"total"`       // 全部
	Normal     int64 `json:"normal"`      // 正常
	LowBalance int64 `json:"low_balance"` // 低余额
	Exhausted  int64 `json:"exhausted"`   // 已用尽（quota=0）
}

// GetUserCreditStats 按「百分比阈值」对用户分类计数。
//
// 已用尽：Quota<=0；低余额：0<剩余且 剩余/基准 < warnPercent%；其余正常。
// 基准取「未过期批次 InitialQuota 之和」；无批次基准时退化为仅按 quota<=0 判定已用尽。
func GetUserCreditStats(warnPercent float64) (*UserCreditStats, error) {
	stats := &UserCreditStats{}
	if err := DB.Model(&User{}).Count(&stats.Total).Error; err != nil {
		return nil, err
	}
	if err := DB.Model(&User{}).Where("quota <= 0").Count(&stats.Exhausted).Error; err != nil {
		return nil, err
	}

	if warnPercent <= 0 {
		stats.Normal = stats.Total - stats.Exhausted
		return stats, nil
	}

	// 低余额需逐用户用「剩余/基准」判定，分页扫描有余额用户
	threshold := warnPercent / 100.0
	low := int64(0)
	lastId := 0
	for {
		var users []User
		if err := DB.Select("id, quota").Where("id > ? AND quota > 0", lastId).
			Order("id ASC").Limit(1000).Find(&users).Error; err != nil {
			return nil, err
		}
		if len(users) == 0 {
			break
		}
		for _, u := range users {
			lastId = u.Id
			base, err := SumActiveInitialQuota(u.Id)
			if err != nil || base <= 0 {
				continue
			}
			if float64(u.Quota)/float64(base) < threshold {
				low++
			}
		}
		if len(users) < 1000 {
			break
		}
	}
	stats.LowBalance = low
	stats.Normal = stats.Total - stats.Exhausted - low
	if stats.Normal < 0 {
		stats.Normal = 0
	}
	return stats, nil
}
