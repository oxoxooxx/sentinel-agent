// Package domain — alert 領域層：domain events
package domain

// AlertDispatched 是當告警成功派送時發出的 domain event
type AlertDispatched struct {
	AlertID int64  // 告警記錄 ID
	Channel string // 派送頻道
	EventID int64  // 觸發此告警的事件 ID
	RuleID  int64  // 觸發此告警的規則 ID
}

// EventName 實作 domain event 識別介面
func (e AlertDispatched) EventName() string {
	return "alert.AlertDispatched"
}
