package controller

import (
	"fmt"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"

	"github.com/gin-gonic/gin"
)

// manageReasonOrDefault 返回备注或默认值。
func manageReasonOrDefault(reason, def string) string {
	if reason == "" {
		return def
	}
	return reason
}

// creditUnitsToQuota 把积分（自定义货币单位）换算为 quota。
func creditUnitsToQuota(units int64) int {
	rate := operation_setting.GetGeneralSetting().CustomCurrencyExchangeRate
	if rate <= 0 {
		rate = 1
	}
	return int(float64(units) / rate * common.QuotaPerUnit)
}

// manageAddQuotaByTier 运营按充值档位代充（image6）：基础积分入 topup 批次，赠送积分入 gift 批次。
// 返回非 nil 表示已写出错误响应，调用方应直接 return。
func manageAddQuotaByTier(c *gin.Context, userId int, req ManageRequest) error {
	tier := operation_setting.FindRechargeTier(req.TierId)
	if tier == nil {
		common.ApiErrorMsg(c, "充值档位不存在")
		return fmt.Errorf("tier not found")
	}
	operatorId := c.GetInt("id")
	reason := manageReasonOrDefault(req.Reason, "管理员按档位代充："+tier.Name)

	// 1) 基础积分（topup 批次）
	baseQuota := creditUnitsToQuota(tier.BaseCredits)
	if baseQuota > 0 {
		if err := model.CreditUserQuota(model.CreditGrant{
			UserId:       userId,
			Source:       model.CreditSourceTopup,
			SourceRef:    "admin_tier:" + tier.Id,
			Quota:        baseQuota,
			ValidityDays: tier.ValidityDays,
			Reason:       reason,
			OperatorId:   operatorId,
			LogType:      model.LogTypeManage,
		}); err != nil {
			common.ApiError(c, err)
			return err
		}
	}

	// 2) 赠送积分（gift 批次）= 档位赠送 + 额外赠送
	giftUnits := tier.GiftCredits
	giftQuota := creditUnitsToQuota(giftUnits)
	if req.GiftValue > 0 {
		giftQuota += req.GiftValue
	}
	if giftQuota > 0 {
		if err := model.CreditUserQuota(model.CreditGrant{
			UserId:       userId,
			Source:       model.CreditSourceGift,
			SourceRef:    "admin_tier:" + tier.Id,
			Quota:        giftQuota,
			ValidityDays: operation_setting.GetCreditSetting().GiftDefaultValidityDays,
			Reason:       reason + "（赠送）",
			OperatorId:   operatorId,
			LogType:      model.LogTypeManage,
		}); err != nil {
			common.ApiError(c, err)
			return err
		}
	}

	common.ApiSuccess(c, gin.H{
		"tier_id":    tier.Id,
		"base_quota": baseQuota,
		"gift_quota": giftQuota,
	})
	return nil
}

// ---------------------------------------------------------------------------
// 充值档位 CRUD（缺口 A-3 / B-17）— 仅超管，写入 payment_setting.recharge_tiers
// ---------------------------------------------------------------------------

// ListRechargeTiers 列出全部充值档位（含下架）。
func ListRechargeTiers(c *gin.Context) {
	common.ApiSuccess(c, operation_setting.GetPaymentSetting().RechargeTiers)
}

func persistRechargeTiers(tiers []operation_setting.RechargeTier) error {
	b, err := common.Marshal(tiers)
	if err != nil {
		return err
	}
	return model.UpdateOption("payment_setting.recharge_tiers", string(b))
}

// SaveRechargeTier 新增或更新一个充值档位（按 Id 匹配；空 Id 视为新增）。
func SaveRechargeTier(c *gin.Context) {
	var tier operation_setting.RechargeTier
	if err := c.ShouldBindJSON(&tier); err != nil {
		common.ApiError(c, err)
		return
	}
	if tier.Name == "" {
		common.ApiErrorMsg(c, "档位名称不能为空")
		return
	}
	if tier.Amount <= 0 {
		common.ApiErrorMsg(c, "充值金额必须大于 0")
		return
	}
	if tier.BaseCredits < 0 || tier.GiftCredits < 0 {
		common.ApiErrorMsg(c, "积分不能为负数")
		return
	}

	tiers := append([]operation_setting.RechargeTier{}, operation_setting.GetPaymentSetting().RechargeTiers...)
	if tier.Id == "" {
		tier.Id = fmt.Sprintf("tier_%d", common.GetTimestamp())
		tiers = append(tiers, tier)
	} else {
		found := false
		for i := range tiers {
			if tiers[i].Id == tier.Id {
				tiers[i] = tier
				found = true
				break
			}
		}
		if !found {
			tiers = append(tiers, tier)
		}
	}
	if err := persistRechargeTiers(tiers); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, tier)
}

// DeleteRechargeTier 删除一个充值档位。
func DeleteRechargeTier(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.ApiErrorMsg(c, "缺少档位 id")
		return
	}
	src := operation_setting.GetPaymentSetting().RechargeTiers
	tiers := make([]operation_setting.RechargeTier, 0, len(src))
	for _, t := range src {
		if t.Id != id {
			tiers = append(tiers, t)
		}
	}
	if err := persistRechargeTiers(tiers); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"id": id})
}

// ---------------------------------------------------------------------------
// 模型价格定时生效计划 CRUD（缺口 A-7）— 仅超管
// ---------------------------------------------------------------------------

type modelPriceScheduleRequest struct {
	ModelName       string  `json:"model_name"`
	ModelRatio      float64 `json:"model_ratio"`
	CompletionRatio float64 `json:"completion_ratio"`
	EffectiveAt     int64   `json:"effective_at"`
}

// ListModelPriceSchedules 列出价格计划，query applied=true/false 过滤。
func ListModelPriceSchedules(c *gin.Context) {
	var applied *bool
	if v := c.Query("applied"); v != "" {
		b := v == "true" || v == "1"
		applied = &b
	}
	list, err := model.ListModelPriceSchedules(applied)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, list)
}

// CreateModelPriceSchedule 新建价格计划。
func CreateModelPriceSchedule(c *gin.Context) {
	var req modelPriceScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	if req.ModelName == "" {
		common.ApiErrorMsg(c, "模型名称不能为空")
		return
	}
	if req.ModelRatio <= 0 && req.CompletionRatio <= 0 {
		common.ApiErrorMsg(c, "至少配置模型倍率或补全倍率之一")
		return
	}
	if req.EffectiveAt <= common.GetTimestamp() {
		common.ApiErrorMsg(c, "生效时间必须晚于当前时间")
		return
	}
	s := &model.ModelPriceSchedule{
		ModelName:       req.ModelName,
		ModelRatio:      req.ModelRatio,
		CompletionRatio: req.CompletionRatio,
		EffectiveAt:     req.EffectiveAt,
		OperatorId:      c.GetInt("id"),
	}
	if err := s.Insert(); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, s)
}

// DeleteModelPriceSchedule 删除一条未生效的价格计划。
func DeleteModelPriceSchedule(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id == 0 {
		common.ApiErrorMsg(c, "缺少计划 id")
		return
	}
	if err := model.DeleteModelPriceSchedule(id); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"id": id})
}
