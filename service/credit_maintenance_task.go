package service

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/billing_setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"

	"github.com/bytedance/gopkg/util/gopool"
)

var (
	creditMaintenanceOnce    sync.Once
	creditMaintenanceRunning atomic.Bool
)

// StartCreditMaintenanceTask 启动积分账本维护后台任务（仅主节点）：
//   - 到期清零（ExpireDueCreditBatches）
//   - 模型价格定时生效（ApplyDueModelPriceSchedules）
//
// 周期取 credit_setting.expire_scan_interval_minutes（默认 5 分钟）。
func StartCreditMaintenanceTask() {
	creditMaintenanceOnce.Do(func() {
		if !common.IsMasterNode {
			return
		}
		gopool.Go(func() {
			interval := time.Duration(operation_setting.GetCreditSetting().GetExpireScanIntervalMinutes()) * time.Minute
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			runCreditMaintenanceOnce()
			for range ticker.C {
				// 周期可能被运营改动，动态对齐
				want := time.Duration(operation_setting.GetCreditSetting().GetExpireScanIntervalMinutes()) * time.Minute
				if want != interval {
					interval = want
					ticker.Reset(interval)
				}
				runCreditMaintenanceOnce()
			}
		})
	})
}

func runCreditMaintenanceOnce() {
	if !creditMaintenanceRunning.CompareAndSwap(false, true) {
		return
	}
	defer creditMaintenanceRunning.Store(false)

	// 0) 老数据回填（仅一次，幂等）：为有余额但无批次的老用户建初始批次
	maybeBackfillCreditBatches()

	// 1) 到期清零：分页处理避免长事务
	for {
		processed, err := model.ExpireDueCreditBatches(200)
		if err != nil {
			common.SysLog("credit maintenance: expire batches error: " + err.Error())
			break
		}
		if processed < 200 {
			break
		}
	}

	// 2) 模型价格定时生效
	applyDueModelPriceSchedules()
}

const creditBackfillDoneOption = "CreditBatchBackfillDone"

// maybeBackfillCreditBatches 首次运行时回填老用户初始批次，用 Option 标记确保只执行一次。
func maybeBackfillCreditBatches() {
	if model.IsCreditBackfillDone() {
		return
	}
	created, err := model.BackfillCreditBatchesForExistingUsers(500)
	if err != nil {
		common.SysLog("credit maintenance: backfill error: " + err.Error())
		return
	}
	common.SysLog(fmt.Sprintf("credit maintenance: backfilled %d credit batches for existing users", created))
	if err := model.UpdateOption(creditBackfillDoneOption, "true"); err != nil {
		common.SysLog("credit maintenance: mark backfill done error: " + err.Error())
	}
}

// applyDueModelPriceSchedules 扫描到点的价格计划并落地。
func applyDueModelPriceSchedules() {
	now := common.GetTimestamp()
	schedules, err := model.GetDueModelPriceSchedules(now, 100)
	if err != nil {
		common.SysLog("credit maintenance: get due price schedules error: " + err.Error())
		return
	}
	if len(schedules) == 0 {
		return
	}

	// 收集所有价格相关 option 的当前值，逐条计划合并进去，最后统一写回。
	patch := newPricingPatch()
	for _, s := range schedules {
		if s.Payload != "" {
			if err := patch.applySnapshot(s.Payload); err != nil {
				common.SysLog(fmt.Sprintf("credit maintenance: apply price snapshot (model %s) error: %s", s.ModelName, err.Error()))
			}
		} else {
			// 旧版兼容：只含 ModelRatio/CompletionRatio 的历史计划
			patch.applyLegacyRatio(s.ModelName, s.ModelRatio, s.CompletionRatio)
		}
	}

	if err := patch.persist(); err != nil {
		common.SysLog("credit maintenance: persist price schedules error: " + err.Error())
		return
	}

	appliedAt := common.GetTimestamp()
	for _, s := range schedules {
		if err := model.MarkModelPriceScheduleApplied(nil, s.Id, appliedAt); err != nil {
			common.SysLog("credit maintenance: mark price schedule applied error: " + err.Error())
		}
	}
}

// modelPriceSnapshot 与前端 ModelRatioData 对齐的价格表单快照。
type modelPriceSnapshot struct {
	Name                 string `json:"name"`
	BillingMode          string `json:"billingMode"`
	Price                string `json:"price"`
	Ratio                string `json:"ratio"`
	CacheRatio           string `json:"cacheRatio"`
	CreateCacheRatio     string `json:"createCacheRatio"`
	CompletionRatio      string `json:"completionRatio"`
	ImageRatio           string `json:"imageRatio"`
	AudioRatio           string `json:"audioRatio"`
	AudioCompletionRatio string `json:"audioCompletionRatio"`
	BillingExpr          string `json:"billingExpr"`
	RequestRuleExpr      string `json:"requestRuleExpr"`
}

// pricingPatch 聚合所有价格 option 的当前映射，支持逐模型合并后统一持久化。
type pricingPatch struct {
	price           map[string]float64
	ratio           map[string]float64
	cache           map[string]float64
	createCache     map[string]float64
	completion      map[string]float64
	image           map[string]float64
	audio           map[string]float64
	audioCompletion map[string]float64
	billingMode     map[string]string
	billingExpr     map[string]string
}

func newPricingPatch() *pricingPatch {
	return &pricingPatch{
		price:           parseFloatMap(ratio_setting.ModelPrice2JSONString()),
		ratio:           parseFloatMap(ratio_setting.ModelRatio2JSONString()),
		cache:           parseFloatMap(ratio_setting.CacheRatio2JSONString()),
		createCache:     parseFloatMap(ratio_setting.CreateCacheRatio2JSONString()),
		completion:      parseFloatMap(ratio_setting.CompletionRatio2JSONString()),
		image:           parseFloatMap(ratio_setting.ImageRatio2JSONString()),
		audio:           parseFloatMap(ratio_setting.AudioRatio2JSONString()),
		audioCompletion: parseFloatMap(ratio_setting.AudioCompletionRatio2JSONString()),
		billingMode:     parseStringMap(billing_setting.GetBillingModeCopy()),
		billingExpr:     parseStringMap(billing_setting.GetBillingExprCopy()),
	}
}

// applySnapshot 把一份价格快照按 billingMode 分支合并进各映射（对齐前端 persistPricingData）。
func (p *pricingPatch) applySnapshot(payloadJSON string) error {
	var snap modelPriceSnapshot
	if err := common.UnmarshalJsonStr(payloadJSON, &snap); err != nil {
		return err
	}
	name := snap.Name
	if name == "" {
		return fmt.Errorf("snapshot missing model name")
	}

	// 先清掉该模型在所有映射里的旧条目，避免模式切换残留。
	delete(p.price, name)
	delete(p.ratio, name)
	delete(p.cache, name)
	delete(p.createCache, name)
	delete(p.completion, name)
	delete(p.image, name)
	delete(p.audio, name)
	delete(p.audioCompletion, name)
	delete(p.billingMode, name)
	delete(p.billingExpr, name)

	setIfPresent := func(m map[string]float64, v string) {
		if v == "" {
			return
		}
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			m[name] = f
		}
	}

	if snap.BillingMode == billing_setting.BillingModeTieredExpr {
		combined := combineBillingExpr(snap.BillingExpr, snap.RequestRuleExpr)
		if combined != "" {
			p.billingMode[name] = billing_setting.BillingModeTieredExpr
			p.billingExpr[name] = combined
		}
		// tiered_expr 同时保留 ratio/price 作为多实例同步延迟期的 fallback。
		setIfPresent(p.price, snap.Price)
		setIfPresent(p.ratio, snap.Ratio)
		setIfPresent(p.cache, snap.CacheRatio)
		setIfPresent(p.createCache, snap.CreateCacheRatio)
		setIfPresent(p.completion, snap.CompletionRatio)
		setIfPresent(p.image, snap.ImageRatio)
		setIfPresent(p.audio, snap.AudioRatio)
		setIfPresent(p.audioCompletion, snap.AudioCompletionRatio)
	} else if snap.Price != "" {
		// 按次计费：只写固定价格
		setIfPresent(p.price, snap.Price)
	} else {
		// 按 token 计费：写各类倍率
		setIfPresent(p.ratio, snap.Ratio)
		setIfPresent(p.cache, snap.CacheRatio)
		setIfPresent(p.createCache, snap.CreateCacheRatio)
		setIfPresent(p.completion, snap.CompletionRatio)
		setIfPresent(p.image, snap.ImageRatio)
		setIfPresent(p.audio, snap.AudioRatio)
		setIfPresent(p.audioCompletion, snap.AudioCompletionRatio)
	}
	return nil
}

// applyLegacyRatio 兼容旧版只含倍率的计划。
func (p *pricingPatch) applyLegacyRatio(name string, modelRatio, completionRatio float64) {
	if name == "" {
		return
	}
	if modelRatio > 0 {
		p.ratio[name] = modelRatio
	}
	if completionRatio > 0 {
		p.completion[name] = completionRatio
	}
}

// persist 把合并后的各映射通过 model.UpdateOption 落地（同时持久化 + 更新内存 + 失效缓存）。
func (p *pricingPatch) persist() error {
	updates := []struct {
		key string
		m   any
	}{
		{"ModelPrice", p.price},
		{"ModelRatio", p.ratio},
		{"CacheRatio", p.cache},
		{"CreateCacheRatio", p.createCache},
		{"CompletionRatio", p.completion},
		{"ImageRatio", p.image},
		{"AudioRatio", p.audio},
		{"AudioCompletionRatio", p.audioCompletion},
		{"billing_setting." + billing_setting.BillingModeField, p.billingMode},
		{"billing_setting." + billing_setting.BillingExprField, p.billingExpr},
	}
	for _, u := range updates {
		out, err := common.Marshal(u.m)
		if err != nil {
			return err
		}
		if err := model.UpdateOption(u.key, string(out)); err != nil {
			return err
		}
	}
	return nil
}

// combineBillingExpr 对齐前端 combineBillingExpr：base 与 requestRule 组合。
func combineBillingExpr(baseExpr, requestRuleExpr string) string {
	base := strings.TrimSpace(baseExpr)
	rules := strings.TrimSpace(requestRuleExpr)
	if base == "" {
		return ""
	}
	if rules == "" {
		return base
	}
	return fmt.Sprintf("(%s) * %s", base, rules)
}

func parseFloatMap(jsonStr string) map[string]float64 {
	m := map[string]float64{}
	if jsonStr != "" {
		_ = common.UnmarshalJsonStr(jsonStr, &m)
	}
	return m
}

func parseStringMap(src map[string]string) map[string]string {
	m := map[string]string{}
	for k, v := range src {
		m[k] = v
	}
	return m
}
