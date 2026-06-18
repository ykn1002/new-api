package service

import (
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/gin-gonic/gin"
)

// SettleCreditConsume 在 relay 结算（已改 User.Quota）之后，按来源优先级把本次实扣 quota
// 从积分批次中扣减（最终一致，异步）。仅钱包计费来源需要分摊；订阅计费不走批次账本。
//
// 返回值无需关心：失败仅记日志，由 ReconcileUserCredit 定时对账兜底。
func SettleCreditConsume(c *gin.Context, relayInfo *relaycommon.RelayInfo, quota int) {
	if relayInfo == nil || quota <= 0 {
		return
	}
	if relayInfo.BillingSource == BillingSourceSubscription {
		return
	}
	if relayInfo.IsPlayground {
		return
	}
	userId := relayInfo.UserId
	gopool.Go(func() {
		deducted, err := model.SettleConsumeToBatches(userId, quota)
		if err != nil {
			common.SysLog("failed to settle credit consume to batches: " + err.Error())
			return
		}
		if c != nil && len(deducted) > 0 {
			logger.LogInfo(c, "credit batch settle: "+formatCreditDeduction(deducted))
		}
	})
}

func formatCreditDeduction(deducted map[string]int) string {
	out := ""
	for src, v := range deducted {
		if out != "" {
			out += ", "
		}
		out += src + "=" + strconv.Itoa(v)
	}
	return out
}
