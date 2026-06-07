// Package domain — event 領域層：domain events
package domain

// EventNormalized 是當原始事件完成正規化後發出的 domain event
// 消費者（如 rule/app/engine.go）訂閱此事件進行規則比對
type EventNormalized struct {
	Event NormalizedEvent // 正規化後的事件
}

// EventName 實作 domain event 識別介面
func (e EventNormalized) EventName() string {
	return "event.EventNormalized"
}
