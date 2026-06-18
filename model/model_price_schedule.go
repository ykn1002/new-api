package model

import (
	"github.com/QuantumNous/new-api/common"

	"gorm.io/gorm"
)

// ModelPriceSchedule 模型价格定时生效计划（缺口 A-7）。
//
// 改价不再立即生效，而是登记一条「待生效」计划，到点由定时任务落地：
// 调 ratio_setting.UpdateModelRatioByJSONString / UpdateCompletionRatioByJSONString。
// 只影响后续请求，不回溯存量。
type ModelPriceSchedule struct {
	Id              int     `json:"id"`
	ModelName       string  `json:"model_name" gorm:"type:varchar(128);index"`
	ModelRatio      float64 `json:"model_ratio"`
	CompletionRatio float64 `json:"completion_ratio"`
	EffectiveAt     int64   `json:"effective_at" gorm:"bigint;index"` // 生效时间戳
	Applied         bool    `json:"applied" gorm:"default:false;index"`
	AppliedAt       int64   `json:"applied_at" gorm:"bigint;default:0"`
	OperatorId      int     `json:"operator_id" gorm:"default:0"`
	CreatedAt       int64   `json:"created_at" gorm:"bigint"`
}

func (ModelPriceSchedule) TableName() string {
	return "model_price_schedules"
}

func (s *ModelPriceSchedule) Insert() error {
	s.CreatedAt = common.GetTimestamp()
	return DB.Create(s).Error
}

func GetModelPriceScheduleById(id int) (*ModelPriceSchedule, error) {
	var s ModelPriceSchedule
	err := DB.Where("id = ?", id).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// ListModelPriceSchedules 列出价格计划，applied 过滤：nil=全部，否则按状态过滤。
func ListModelPriceSchedules(applied *bool) ([]ModelPriceSchedule, error) {
	var list []ModelPriceSchedule
	tx := DB.Model(&ModelPriceSchedule{})
	if applied != nil {
		tx = tx.Where("applied = ?", *applied)
	}
	err := tx.Order("effective_at ASC, id ASC").Find(&list).Error
	return list, err
}

func DeleteModelPriceSchedule(id int) error {
	return DB.Where("id = ?", id).Delete(&ModelPriceSchedule{}).Error
}

// GetDueModelPriceSchedules 取到点且未生效的计划（effective_at<=now && !applied）。
func GetDueModelPriceSchedules(now int64, limit int) ([]ModelPriceSchedule, error) {
	if limit <= 0 {
		limit = 100
	}
	var list []ModelPriceSchedule
	err := DB.Where("applied = ? AND effective_at <= ?", false, now).
		Order("effective_at ASC, id ASC").Limit(limit).Find(&list).Error
	return list, err
}

// MarkModelPriceScheduleApplied 标记计划已生效。
func MarkModelPriceScheduleApplied(tx *gorm.DB, id int, appliedAt int64) error {
	db := DB
	if tx != nil {
		db = tx
	}
	return db.Model(&ModelPriceSchedule{}).Where("id = ?", id).
		Updates(map[string]interface{}{"applied": true, "applied_at": appliedAt}).Error
}
