// Package domain — source 領域層：domain events
package domain

// SourceRegistered 是當新事件來源成功建立時發出的 domain event
type SourceRegistered struct {
	Source Source // 新建立的事件來源
}

// EventName 實作 domain event 識別介面
func (e SourceRegistered) EventName() string {
	return "source.SourceRegistered"
}
