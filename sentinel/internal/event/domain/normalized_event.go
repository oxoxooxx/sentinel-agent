// Package domain — event 領域層：NormalizedEvent aggregate
package domain

import "time"

// NormalizedEvent 代表經過正規化的安全事件（aggregate root）
// 由 ACL 解碼器將原始 log 轉換而來，供規則引擎與 Dashboard 使用
type NormalizedEvent struct {
	ID         int64     // 資料庫主鍵
	SourceID   int64     // 對應的事件來源 ID
	RawLog     string    // 原始 log 內容
	NormalAt   time.Time // 正規化完成時間
	Severity   Severity  // 嚴重等級
	Category   string    // 分類，如 "firewall", "nas", "webhook"
	Message    string    // 人類可讀摘要
	ExtraJSON  string    // 額外結構化資料（JSON 字串）
}

// NewNormalizedEvent 建立 NormalizedEvent，確保必要欄位不為空
func NewNormalizedEvent(sourceID int64, rawLog string, sev Severity, category, message string) NormalizedEvent {
	return NormalizedEvent{
		SourceID:  sourceID,
		RawLog:    rawLog,
		NormalAt:  time.Now(),
		Severity:  sev,
		Category:  category,
		Message:   message,
		ExtraJSON: "{}",
	}
}

// WithExtra 附加額外 JSON 資料，回傳新物件（immutable pattern）
func (e NormalizedEvent) WithExtra(extraJSON string) NormalizedEvent {
	updated := e
	if extraJSON == "" {
		updated.ExtraJSON = "{}"
	} else {
		updated.ExtraJSON = extraJSON
	}
	return updated
}

// NeedsAlert 判斷此事件是否需要觸發告警
func (e NormalizedEvent) NeedsAlert() bool {
	return e.Severity.IsAlertable()
}
