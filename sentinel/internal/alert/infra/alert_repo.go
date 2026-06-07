// Package infra — alert 基礎設施層：告警 DB 存取
package infra

import (
	"context"
	"fmt"

	alertdomain "github.com/oxoxooxx/sentinel/internal/alert/domain"
	eventinfra "github.com/oxoxooxx/sentinel/internal/event/infra"
)

// AlertRepository 告警資料存取物件
type AlertRepository struct {
	db eventinfra.DB
}

// NewAlertRepository 建立 AlertRepository
func NewAlertRepository(db eventinfra.DB) *AlertRepository {
	return &AlertRepository{db: db}
}

// Save 儲存一筆新告警
func (r *AlertRepository) Save(ctx context.Context, alert alertdomain.Alert) (alertdomain.Alert, error) {
	dbAlert := eventinfra.Alert{
		EventID:   alert.EventID,
		RuleID:    alert.RuleID,
		Status:    eventinfra.AlertStatus(alert.Status),
		Channel:   alert.Channel,
		Message:   alert.Message,
		CreatedAt: alert.CreatedAt,
		SentAt:    alert.SentAt,
		AckedAt:   alert.AckedAt,
	}

	saved, err := r.db.SaveAlert(ctx, dbAlert)
	if err != nil {
		return alertdomain.Alert{}, fmt.Errorf("儲存告警失敗: %w", err)
	}

	alert.ID = saved.ID
	return alert, nil
}

// UpdateStatus 更新告警狀態
func (r *AlertRepository) UpdateStatus(ctx context.Context, id int64, status alertdomain.AlertStatus) error {
	if err := r.db.UpdateAlertStatus(ctx, id, eventinfra.AlertStatus(status)); err != nil {
		return fmt.Errorf("更新告警狀態失敗 id=%d: %w", id, err)
	}
	return nil
}
