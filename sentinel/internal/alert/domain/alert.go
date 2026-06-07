// Package domain — alert 領域層：Alert aggregate
package domain

import "time"

// AlertStatus 告警狀態
type AlertStatus string

const (
	AlertStatusPending  AlertStatus = "pending"  // 尚未處理
	AlertStatusSent     AlertStatus = "sent"     // 已送出通知
	AlertStatusAcked    AlertStatus = "acked"    // 已確認
	AlertStatusResolved AlertStatus = "resolved" // 已解決
)

// Alert 代表一筆告警記錄（aggregate root）
type Alert struct {
	ID        int64
	EventID   int64       // 觸發此告警的事件
	RuleID    int64       // 觸發此告警的規則
	Status    AlertStatus
	Channel   string     // 通知頻道，如 "line", "telegram"
	Message   string     // 送出的告警訊息
	CreatedAt time.Time
	SentAt    *time.Time
	AckedAt   *time.Time
}

// NewAlert 建立一筆新告警
func NewAlert(eventID, ruleID int64, channel, message string) Alert {
	return Alert{
		EventID:   eventID,
		RuleID:    ruleID,
		Status:    AlertStatusPending,
		Channel:   channel,
		Message:   message,
		CreatedAt: time.Now(),
	}
}

// MarkSent 標記告警為已送出，回傳新物件（immutable pattern）
func (a Alert) MarkSent() Alert {
	now := time.Now()
	updated := a
	updated.Status = AlertStatusSent
	updated.SentAt = &now
	return updated
}

// Ack 確認告警，回傳新物件（immutable pattern）
func (a Alert) Ack() Alert {
	now := time.Now()
	updated := a
	updated.Status = AlertStatusAcked
	updated.AckedAt = &now
	return updated
}

// Resolve 解決告警，回傳新物件（immutable pattern）
func (a Alert) Resolve() Alert {
	updated := a
	updated.Status = AlertStatusResolved
	return updated
}

// IsPending 檢查告警是否尚未處理
func (a Alert) IsPending() bool {
	return a.Status == AlertStatusPending
}
