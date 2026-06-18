package service

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
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
	modelRatios := map[string]float64{}
	completionRatios := map[string]float64{}
	for _, s := range schedules {
		if s.ModelRatio > 0 {
			modelRatios[s.ModelName] = s.ModelRatio
		}
		if s.CompletionRatio > 0 {
			completionRatios[s.ModelName] = s.CompletionRatio
		}
	}
	if len(modelRatios) > 0 {
		if err := applyRatioPatch(ratio_setting.ModelRatio2JSONString, ratio_setting.UpdateModelRatioByJSONString, modelRatios); err != nil {
			common.SysLog("credit maintenance: apply model ratio error: " + err.Error())
			return
		}
	}
	if len(completionRatios) > 0 {
		if err := applyRatioPatch(ratio_setting.CompletionRatio2JSONString, ratio_setting.UpdateCompletionRatioByJSONString, completionRatios); err != nil {
			common.SysLog("credit maintenance: apply completion ratio error: " + err.Error())
			return
		}
	}
	appliedAt := common.GetTimestamp()
	for _, s := range schedules {
		if err := model.MarkModelPriceScheduleApplied(nil, s.Id, appliedAt); err != nil {
			common.SysLog("credit maintenance: mark price schedule applied error: " + err.Error())
		}
	}
	// 持久化到 Option 表，保证重启后生效
	persistRatioOption("ModelRatio", ratio_setting.ModelRatio2JSONString())
	persistRatioOption("CompletionRatio", ratio_setting.CompletionRatio2JSONString())
}

// applyRatioPatch 读取当前倍率 JSON，合并 patch 后写回内存映射。
func applyRatioPatch(currentJSON func() string, update func(string) error, patch map[string]float64) error {
	merged, err := mergeRatioJSON(currentJSON(), patch)
	if err != nil {
		return err
	}
	return update(merged)
}

func mergeRatioJSON(currentJSON string, patch map[string]float64) (string, error) {
	current := map[string]float64{}
	if currentJSON != "" {
		if err := common.UnmarshalJsonStr(currentJSON, &current); err != nil {
			return "", err
		}
	}
	for k, v := range patch {
		current[k] = v
	}
	out, err := common.Marshal(current)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func persistRatioOption(key, value string) {
	if err := model.UpdateOption(key, value); err != nil {
		common.SysLog("credit maintenance: persist " + key + " error: " + err.Error())
	}
}
