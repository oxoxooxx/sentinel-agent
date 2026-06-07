// Package domain — ingestion 領域層：domain events
package domain

// RawEventReceived 是當新原始事件進入系統時發出的 domain event
// 消費者（如 event/app/normalize_service.go）訂閱此事件進行正規化
type RawEventReceived struct {
	EndpointID string   // 觸發此事件的端點 ID
	RawEvent   RawEvent // 接收到的原始事件
}

// EventName 實作 domain event 識別介面
func (e RawEventReceived) EventName() string {
	return "ingestion.RawEventReceived"
}
