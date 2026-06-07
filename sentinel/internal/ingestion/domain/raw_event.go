// Package domain — ingestion 領域層：RawEvent value object
package domain

import "time"

// RawEvent 代表從外部接收到的原始事件（value object，不可變）
// 此為接收層的原始資料，尚未正規化
type RawEvent struct {
	Source     string    // 來源識別（如 IP 位址、webhook source）
	Protocol   string    // 傳輸協議（"udp", "http"）
	RawData    string    // 原始資料內容
	ReceivedAt time.Time // 接收時間
}

// NewRawEvent 建立 RawEvent，確保 ReceivedAt 不為零值
func NewRawEvent(source, protocol, rawData string) RawEvent {
	return RawEvent{
		Source:     source,
		Protocol:   protocol,
		RawData:    rawData,
		ReceivedAt: time.Now(),
	}
}

// IsEmpty 檢查原始資料是否為空
func (e RawEvent) IsEmpty() bool {
	return e.RawData == ""
}
