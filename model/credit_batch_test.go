package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/require"
)

func clearCreditTables(t *testing.T) {
	t.Helper()
	require.NoError(t, DB.Exec("DELETE FROM credit_batches").Error)
	require.NoError(t, DB.Exec("DELETE FROM users WHERE id >= 7000 AND id < 8000").Error)
	t.Cleanup(func() {
		DB.Exec("DELETE FROM credit_batches")
		DB.Exec("DELETE FROM users WHERE id >= 7000 AND id < 8000")
	})
}

func TestCreditUserQuotaInvariant(t *testing.T) {
	clearCreditTables(t)
	common.BatchUpdateEnabled = false

	u := &User{Id: 7001, Username: "credit_u1", Quota: 0, AffCode: "cb7001"}
	require.NoError(t, DB.Create(u).Error)

	require.NoError(t, CreditUserQuota(CreditGrant{
		UserId: u.Id, Source: CreditSourceTopup, Quota: 1000, ValidityDays: 0,
	}))
	require.NoError(t, CreditUserQuota(CreditGrant{
		UserId: u.Id, Source: CreditSourceGift, Quota: 500, ValidityDays: 7,
	}))

	// 不变量 I：Σremaining == User.Quota
	var refreshed User
	require.NoError(t, DB.First(&refreshed, u.Id).Error)
	require.Equal(t, 1500, refreshed.Quota)

	summaries, err := GetCreditSourceSummaries(u.Id)
	require.NoError(t, err)
	sum := 0
	for _, s := range summaries {
		sum += s.Remaining
	}
	require.Equal(t, refreshed.Quota, sum)
}

func TestSettleConsumeToBatchesOrdering(t *testing.T) {
	clearCreditTables(t)
	common.BatchUpdateEnabled = false

	userId := 7002
	require.NoError(t, DB.Create(&User{Id: userId, Username: "credit_u2", Quota: 0, AffCode: "cb7002"}).Error)

	// 建三个批次：topup(永久) / subscription(到期晚) / gift(到期早)
	require.NoError(t, CreditUserQuota(CreditGrant{UserId: userId, Source: CreditSourceTopup, Quota: 300}))
	require.NoError(t, CreditUserQuota(CreditGrant{UserId: userId, Source: CreditSourceSubscription, Quota: 200, ValidityDays: 30}))
	require.NoError(t, CreditUserQuota(CreditGrant{UserId: userId, Source: CreditSourceGift, Quota: 100, ValidityDays: 7}))

	// 扣 250：应先扣 gift(100) 再扣 subscription(150)，topup 不动
	deducted, err := SettleConsumeToBatches(userId, 250)
	require.NoError(t, err)
	require.Equal(t, 100, deducted[CreditSourceGift])
	require.Equal(t, 150, deducted[CreditSourceSubscription])
	require.Equal(t, 0, deducted[CreditSourceTopup])

	// 主来源应为 subscription（占比最大）
	require.Equal(t, CreditSourceSubscription, PrimaryCreditSource(deducted))

	summaries, err := GetCreditSourceSummaries(userId)
	require.NoError(t, err)
	remainBySource := map[string]int{}
	for _, s := range summaries {
		remainBySource[s.Source] = s.Remaining
	}
	require.Equal(t, 50, remainBySource[CreditSourceSubscription])
	require.Equal(t, 300, remainBySource[CreditSourceTopup])
	_, giftExists := remainBySource[CreditSourceGift]
	require.False(t, giftExists) // gift 已扣尽
}

func TestReconcileUserCredit(t *testing.T) {
	clearCreditTables(t)
	common.BatchUpdateEnabled = false

	userId := 7003
	require.NoError(t, DB.Create(&User{Id: userId, Username: "credit_u3", Quota: 0, AffCode: "cb7003"}).Error)
	require.NoError(t, CreditUserQuota(CreditGrant{UserId: userId, Source: CreditSourceTopup, Quota: 1000}))

	// 人为制造漂移：直接把 User.Quota 改大（模拟绕过账本的扣加）
	require.NoError(t, DB.Model(&User{}).Where("id = ?", userId).Update("quota", 1200).Error)

	drifted, delta, err := ReconcileUserCredit(userId)
	require.NoError(t, err)
	require.True(t, drifted)
	require.Equal(t, 200, delta)

	// 修正后 Σremaining 应等于 User.Quota
	base, err := SumActiveInitialQuota(userId)
	require.NoError(t, err)
	require.Equal(t, 1200, base)
}
