package model

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/QuantumNous/new-api/common"

	"gorm.io/gorm"
)

// CreditBatch 积分批次账本（平行子账本）。
//
// 设计要点（详见 docs/Token-Plan-技术方案.md 第 2 节）：
//   - User.Quota 仍是唯一权威总余额；批次账本是派生明细。
//   - 不变量 I：某用户所有「未过期且 remaining>0」批次的 remaining 之和 == User.Quota。
//   - InitialQuota/Remaining 直接存 quota（int，与 User.Quota 同口径），对外再投影成积分。
//   - 扣减优先级：gift → subscription → topup；同类按 ExpireAt 升序（永久排最后）。
type CreditBatch struct {
	Id           int    `json:"id"`
	UserId       int    `json:"user_id" gorm:"index:idx_cb_user_status,priority:1;not null"`
	Source       string `json:"source" gorm:"type:varchar(16);index;not null"` // gift/subscription/topup
	SourceRef    string `json:"source_ref" gorm:"type:varchar(64);default:''"` // 订单号/套餐周期/操作单号
	InitialQuota int    `json:"initial_quota" gorm:"not null"`
	Remaining    int    `json:"remaining" gorm:"index:idx_cb_user_status,priority:2;not null"`
	Status       int    `json:"status" gorm:"default:1;index:idx_cb_user_status,priority:3"` // 1=active 2=exhausted 3=expired
	ExpireAt     int64  `json:"expire_at" gorm:"bigint;index:idx_cb_expire"`                // 到期时间戳；0=永久
	CreatedAt    int64  `json:"created_at" gorm:"bigint"`
	UpdatedAt    int64  `json:"updated_at" gorm:"bigint"`
}

// 批次状态
const (
	CreditBatchStatusActive    = 1
	CreditBatchStatusExhausted = 2
	CreditBatchStatusExpired   = 3
)

func (CreditBatch) TableName() string {
	return "credit_batches"
}

// sourcePriority 返回来源扣减优先级（数值小者先扣）：gift < subscription < topup。
func sourcePriority(source string) int {
	switch source {
	case CreditSourceGift:
		return 0
	case CreditSourceSubscription:
		return 1
	case CreditSourceTopup:
		return 2
	default:
		return 3
	}
}

// creditBatchUserLocks 单用户批次写操作串行化（进程内）。
// 跨节点一致性由「以 User.Quota 为准的对账」兜底（ReconcileUserCredit）。
var creditBatchUserLocks sync.Map

func lockCreditBatchUser(userId int) func() {
	v, _ := creditBatchUserLocks.LoadOrStore(userId, &sync.Mutex{})
	mu := v.(*sync.Mutex)
	mu.Lock()
	return func() { mu.Unlock() }
}

// CreateCreditBatchTx 在事务内插入一个积分批次（供 CreditUserQuota 使用）。
func CreateCreditBatchTx(tx *gorm.DB, batch *CreditBatch) error {
	if batch.UserId == 0 {
		return errors.New("user id is required")
	}
	if batch.InitialQuota <= 0 {
		return errors.New("initial quota must be positive")
	}
	now := common.GetTimestamp()
	batch.CreatedAt = now
	batch.UpdatedAt = now
	if batch.Remaining == 0 {
		batch.Remaining = batch.InitialQuota
	}
	if batch.Status == 0 {
		batch.Status = CreditBatchStatusActive
	}
	return tx.Create(batch).Error
}

// activeBatchesForUpdate 取用户当前可扣减批次（status=active && remaining>0），按扣减优先级排序。
// 调用方应处于事务且已持有用户级串行锁。
func activeBatchesForUpdate(tx *gorm.DB, userId int) ([]CreditBatch, error) {
	var batches []CreditBatch
	err := tx.Where("user_id = ? AND status = ? AND remaining > 0", userId, CreditBatchStatusActive).
		Find(&batches).Error
	if err != nil {
		return nil, err
	}
	sortBatchesByDeductOrder(batches)
	return batches, nil
}

func sortBatchesByDeductOrder(batches []CreditBatch) {
	sort.SliceStable(batches, func(i, j int) bool {
		pi, pj := sourcePriority(batches[i].Source), sourcePriority(batches[j].Source)
		if pi != pj {
			return pi < pj
		}
		// 永久（ExpireAt=0）排最后
		ei, ej := batches[i].ExpireAt, batches[j].ExpireAt
		iPerm, jPerm := ei == 0, ej == 0
		if iPerm != jPerm {
			return !iPerm
		}
		if ei != ej {
			return ei < ej
		}
		return batches[i].Id < batches[j].Id
	})
}

// SettleConsumeToBatches 按优先级把一笔已扣 quota 从批次 Remaining 中扣掉。
//
// 在 relay 结算改完 User.Quota 之后调用（可异步，最终一致）。返回各来源实际扣减量，
// 供调用方写「来源 + 操作后余额」流水。amount<=0 时不处理。
func SettleConsumeToBatches(userId int, amount int) (map[string]int, error) {
	result := map[string]int{}
	if amount <= 0 {
		return result, nil
	}
	unlock := lockCreditBatchUser(userId)
	defer unlock()

	err := DB.Transaction(func(tx *gorm.DB) error {
		batches, err := activeBatchesForUpdate(tx, userId)
		if err != nil {
			return err
		}
		remaining := amount
		now := common.GetTimestamp()
		for i := range batches {
			if remaining <= 0 {
				break
			}
			b := &batches[i]
			deduct := b.Remaining
			if deduct > remaining {
				deduct = remaining
			}
			newRemaining := b.Remaining - deduct
			newStatus := b.Status
			if newRemaining == 0 {
				newStatus = CreditBatchStatusExhausted
			}
			if err := tx.Model(&CreditBatch{}).Where("id = ?", b.Id).
				Updates(map[string]interface{}{
					"remaining":  newRemaining,
					"status":     newStatus,
					"updated_at": now,
				}).Error; err != nil {
				return err
			}
			result[b.Source] += deduct
			remaining -= deduct
		}
		// remaining>0 表示批次不足以覆盖本次消费（账本滞后/漂移），不阻断主流程，
		// 交由 ReconcileUserCredit 以 User.Quota 为准修正。
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// PrimaryCreditSource 返回一次扣减中占比最大的来源（用于流水 CreditSource 列）。
func PrimaryCreditSource(deducted map[string]int) string {
	best := ""
	bestVal := 0
	for src, v := range deducted {
		if v > bestVal {
			bestVal = v
			best = src
		}
	}
	return best
}

// CreditSourceSummary 某用户按来源聚合的可用余额（status=active && remaining>0）。
type CreditSourceSummary struct {
	Source       string `json:"source"`
	Remaining    int    `json:"remaining"`
	NearestExpire int64 `json:"nearest_expire"` // 该来源下最近到期批次的 expire_at；0=含永久
}

// GetCreditSourceSummaries 按来源聚合用户当前可用余额及最近到期时间。
func GetCreditSourceSummaries(userId int) ([]CreditSourceSummary, error) {
	var rows []struct {
		Source    string
		Remaining int
	}
	err := DB.Model(&CreditBatch{}).
		Select("source, SUM(remaining) as remaining").
		Where("user_id = ? AND status = ? AND remaining > 0", userId, CreditBatchStatusActive).
		Group("source").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	summaries := make([]CreditSourceSummary, 0, len(rows))
	for _, r := range rows {
		s := CreditSourceSummary{Source: r.Source, Remaining: r.Remaining}
		// 取该来源下最近到期（expire_at>0）批次时间
		var nearest int64
		DB.Model(&CreditBatch{}).
			Select("expire_at").
			Where("user_id = ? AND source = ? AND status = ? AND remaining > 0 AND expire_at > 0",
				userId, r.Source, CreditBatchStatusActive).
			Order("expire_at ASC").Limit(1).Scan(&nearest)
		s.NearestExpire = nearest
		summaries = append(summaries, s)
	}
	return summaries, nil
}

// NearestExpiringBatch 即将过期提示：取用户最近一笔将到期且 remaining>0 的批次。
func NearestExpiringBatch(userId int) (*CreditBatch, error) {
	var batch CreditBatch
	err := DB.Where("user_id = ? AND status = ? AND remaining > 0 AND expire_at > 0",
		userId, CreditBatchStatusActive).
		Order("expire_at ASC").First(&batch).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &batch, nil
}

// GetUserCreditBatches 返回用户全部批次（用于明细/管理展示），按创建时间倒序。
func GetUserCreditBatches(userId int) ([]CreditBatch, error) {
	var batches []CreditBatch
	err := DB.Where("user_id = ?", userId).Order("created_at DESC, id DESC").Find(&batches).Error
	return batches, err
}

// SumActiveInitialQuota 返回用户未过期批次 InitialQuota 之和（百分比预警基准）。
func SumActiveInitialQuota(userId int) (int, error) {
	var total int64
	err := DB.Model(&CreditBatch{}).
		Where("user_id = ? AND status = ?", userId, CreditBatchStatusActive).
		Select("COALESCE(SUM(initial_quota),0)").Scan(&total).Error
	if err != nil {
		return 0, err
	}
	return int(total), nil
}

// ExpireDueCreditBatches 到期清零：把 expire_at!=0 && expire_at<now && status=active 的批次
// Remaining 清零并同额扣减 User.Quota，写「过期」流水。分页限流处理，返回处理条数。
func ExpireDueCreditBatches(batchSize int) (int, error) {
	if batchSize <= 0 {
		batchSize = 200
	}
	now := common.GetTimestamp()
	var due []CreditBatch
	err := DB.Where("status = ? AND expire_at != 0 AND expire_at < ?", CreditBatchStatusActive, now).
		Order("user_id ASC, id ASC").Limit(batchSize).Find(&due).Error
	if err != nil {
		return 0, err
	}
	processed := 0
	for i := range due {
		b := due[i]
		if err := expireOneBatch(b); err != nil {
			common.SysLog(fmt.Sprintf("failed to expire credit batch %d (user %d): %s", b.Id, b.UserId, err.Error()))
			continue
		}
		processed++
	}
	return processed, nil
}

func expireOneBatch(b CreditBatch) error {
	unlock := lockCreditBatchUser(b.UserId)
	defer unlock()

	var freed int
	err := DB.Transaction(func(tx *gorm.DB) error {
		// 重新读取，确保仍 active 且 remaining 未被并发改动
		var cur CreditBatch
		if err := tx.Where("id = ?", b.Id).First(&cur).Error; err != nil {
			return err
		}
		if cur.Status != CreditBatchStatusActive {
			return nil
		}
		freed = cur.Remaining
		now := common.GetTimestamp()
		if err := tx.Model(&CreditBatch{}).Where("id = ?", cur.Id).
			Updates(map[string]interface{}{
				"remaining":  0,
				"status":     CreditBatchStatusExpired,
				"updated_at": now,
			}).Error; err != nil {
			return err
		}
		if freed > 0 {
			if err := tx.Model(&User{}).Where("id = ?", cur.UserId).
				Update("quota", gorm.Expr("quota - ?", freed)).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if freed > 0 {
		// 同步缓存余额
		_ = cacheDecrUserQuota(b.UserId, int64(freed))
		RecordLog(b.UserId, LogTypeExpire, fmt.Sprintf("积分批次到期清零，来源 %s，过期额度 %d", b.Source, freed))
	}
	return nil
}

// ReconcileUserCredit 对账兜底：校验 Σremaining == User.Quota，偏差时以 User.Quota 为准修正批次。
//
// 修正策略：
//   - Σremaining > Quota：从优先级靠后的批次按序扣减差额（多记的回收）。
//   - Σremaining < Quota：把差额补到一个 source=topup 的永久批次（少记的补足）。
//
// 返回 (是否有偏差, 偏差量 deltaQuota=Quota-Σremaining, error)。
func ReconcileUserCredit(userId int) (bool, int, error) {
	unlock := lockCreditBatchUser(userId)
	defer unlock()

	var drifted bool
	var delta int
	err := DB.Transaction(func(tx *gorm.DB) error {
		var user User
		if err := tx.Where("id = ?", userId).First(&user).Error; err != nil {
			return err
		}
		batches, err := activeBatchesForUpdate(tx, userId)
		if err != nil {
			return err
		}
		sum := 0
		for _, b := range batches {
			sum += b.Remaining
		}
		delta = user.Quota - sum
		if delta == 0 {
			return nil
		}
		drifted = true
		now := common.GetTimestamp()
		if delta < 0 {
			// 批次多记，从优先级靠后的批次倒序回收 -delta
			need := -delta
			for i := len(batches) - 1; i >= 0 && need > 0; i-- {
				b := &batches[i]
				cut := b.Remaining
				if cut > need {
					cut = need
				}
				newRemaining := b.Remaining - cut
				newStatus := b.Status
				if newRemaining == 0 {
					newStatus = CreditBatchStatusExhausted
				}
				if err := tx.Model(&CreditBatch{}).Where("id = ?", b.Id).
					Updates(map[string]interface{}{"remaining": newRemaining, "status": newStatus, "updated_at": now}).Error; err != nil {
					return err
				}
				need -= cut
			}
		} else {
			// 批次少记，补一个永久 topup 批次
			fix := &CreditBatch{
				UserId:       userId,
				Source:       CreditSourceTopup,
				SourceRef:    "reconcile",
				InitialQuota: delta,
				Remaining:    delta,
				Status:       CreditBatchStatusActive,
				ExpireAt:     0,
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if err := tx.Create(fix).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return false, 0, err
	}
	if drifted {
		common.SysLog(fmt.Sprintf("ReconcileUserCredit: user %d drift corrected, delta(Quota-Σremaining)=%d", userId, delta))
	}
	return drifted, delta, nil
}

// BackfillCreditBatchesForExistingUsers 上线回填：为「有余额但无任何批次」的老用户生成一个
// 永久 topup 初始批次，令不变量 I 初始成立。分页处理，幂等（已存在批次的用户跳过）。
func BackfillCreditBatchesForExistingUsers(pageSize int) (int, error) {
	if pageSize <= 0 {
		pageSize = 500
	}
	created := 0
	lastId := 0
	for {
		var users []User
		err := DB.Where("id > ? AND quota > 0", lastId).
			Order("id ASC").Limit(pageSize).Find(&users).Error
		if err != nil {
			return created, err
		}
		if len(users) == 0 {
			break
		}
		for _, u := range users {
			lastId = u.Id
			var cnt int64
			if err := DB.Model(&CreditBatch{}).Where("user_id = ?", u.Id).Count(&cnt).Error; err != nil {
				return created, err
			}
			if cnt > 0 {
				continue
			}
			now := common.GetTimestamp()
			batch := &CreditBatch{
				UserId:       u.Id,
				Source:       CreditSourceTopup,
				SourceRef:    "backfill",
				InitialQuota: u.Quota,
				Remaining:    u.Quota,
				Status:       CreditBatchStatusActive,
				ExpireAt:     0,
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if err := DB.Create(batch).Error; err != nil {
				common.SysLog(fmt.Sprintf("backfill credit batch failed for user %d: %s", u.Id, err.Error()))
				continue
			}
			created++
		}
		if len(users) < pageSize {
			break
		}
	}
	return created, nil
}

// CreditGrant 描述一次「加积分」进账。
type CreditGrant struct {
	UserId       int
	Source       string // gift/subscription/topup
	SourceRef    string
	Quota        int    // 入账 quota（由积分/金额 × 汇率换算得到）
	ValidityDays int    // 0=永久
	Reason       string // 流水备注
	OperatorId   int    // 运营代充时记录操作人；0=系统/用户自助
	LogType      int    // 流水类型，0 时按来源推断（topup→Topup，其余→Manage）
}

// CreditUserQuota 所有「加积分」的唯一入口：同一事务内建批次 + 加 User.Quota，事务后写流水。
//
//   - 保证不变量 I：批次 Remaining 之和与 User.Quota 同步增加。
//   - 复用现有缓存：事务提交后同步 Redis 缓存余额（与 IncreaseUserQuota 一致）。
func CreditUserQuota(g CreditGrant) error {
	if g.Quota <= 0 {
		return errors.New("credit quota must be positive")
	}
	if g.Source == "" {
		g.Source = CreditSourceTopup
	}
	unlock := lockCreditBatchUser(g.UserId)
	defer unlock()

	var expireAt int64
	if g.ValidityDays > 0 {
		expireAt = common.GetTimestamp() + int64(g.ValidityDays)*86400
	}

	err := DB.Transaction(func(tx *gorm.DB) error {
		batch := &CreditBatch{
			UserId:       g.UserId,
			Source:       g.Source,
			SourceRef:    g.SourceRef,
			InitialQuota: g.Quota,
			Remaining:    g.Quota,
			Status:       CreditBatchStatusActive,
			ExpireAt:     expireAt,
		}
		if err := CreateCreditBatchTx(tx, batch); err != nil {
			return err
		}
		if err := tx.Model(&User{}).Where("id = ?", g.UserId).
			Update("quota", gorm.Expr("quota + ?", g.Quota)).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	// 同步缓存余额（与 IncreaseUserQuota 一致，失败仅记日志）
	if cacheErr := cacheIncrUserQuota(g.UserId, int64(g.Quota)); cacheErr != nil {
		common.SysLog("failed to incr user quota cache after credit: " + cacheErr.Error())
	}

	// 写流水（含来源 + 操作后余额）
	logType := g.LogType
	if logType == 0 {
		if g.Source == CreditSourceTopup {
			logType = LogTypeTopup
		} else {
			logType = LogTypeManage
		}
	}
	balanceAfter, _ := GetUserQuota(g.UserId, true)
	content := g.Reason
	if content == "" {
		content = fmt.Sprintf("积分入账，来源 %s，额度 %d", g.Source, g.Quota)
	}
	RecordCreditLog(g.UserId, logType, content, g.Source, balanceAfter)
	return nil
}

// RegisterTopupCreditBatch 为「已通过其它路径加到 User.Quota 的充值额度」补登一个 topup 批次，
// 不再改动 User.Quota（仅维护批次账本，保持不变量 I）。validityDays<0 时用默认充值有效期。
//
// 用于已有支付回调（epay/stripe/creem/waffo）——它们在事务内已直接 quota+=amount，
// 这里只补账本明细 + 流水来源/操作后余额。
func RegisterTopupCreditBatch(userId int, quota int, sourceRef string, validityDays int) error {
	if quota <= 0 {
		return nil
	}
	unlock := lockCreditBatchUser(userId)
	defer unlock()

	var expireAt int64
	if validityDays > 0 {
		expireAt = common.GetTimestamp() + int64(validityDays)*86400
	}
	batch := &CreditBatch{
		UserId:       userId,
		Source:       CreditSourceTopup,
		SourceRef:    sourceRef,
		InitialQuota: quota,
		Remaining:    quota,
		Status:       CreditBatchStatusActive,
		ExpireAt:     expireAt,
	}
	if err := DB.Transaction(func(tx *gorm.DB) error {
		return CreateCreditBatchTx(tx, batch)
	}); err != nil {
		return err
	}
	balanceAfter, _ := GetUserQuota(userId, true)
	RecordCreditLog(userId, LogTypeTopup, fmt.Sprintf("充值到账，额度 %d", quota), CreditSourceTopup, balanceAfter)
	return nil
}

// IsCreditBackfillDone 检查老数据回填是否已执行过（Option 标记）。
func IsCreditBackfillDone() bool {
	var opt Option
	err := DB.Where(&Option{Key: "CreditBatchBackfillDone"}).First(&opt).Error
	if err != nil {
		return false
	}
	return opt.Value == "true"
}

