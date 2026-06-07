// Package domain — rule 領域層：Condition value object
package domain

// ConditionType 規則條件類型
type ConditionType string

const (
	ConditionTypeKeyword   ConditionType = "keyword"   // 關鍵字比對
	ConditionTypeThreshold ConditionType = "threshold" // 閾值觸發
	ConditionTypeWhitelist ConditionType = "whitelist" // 白名單排除
)

// Condition 規則條件（value object，不可變）
// 對應 rules.cond_json 欄位的反序列化結構
type Condition struct {
	Type      ConditionType `json:"type"`
	Field     string        `json:"field"`               // 比對的欄位（"message", "raw_log", "severity" 等）
	Keywords  []string      `json:"keywords,omitempty"`  // type=keyword 時使用
	Threshold int           `json:"threshold,omitempty"` // type=threshold 時使用（N 分鐘內 N 筆）
	WindowMin int           `json:"window_min,omitempty"` // 閾值時間窗口（分鐘）
	Whitelist []string      `json:"whitelist,omitempty"` // type=whitelist 時使用
	Severity  string        `json:"severity,omitempty"`  // 觸發告警的嚴重等級
}

// IsValid 檢查條件是否合法
func (c Condition) IsValid() bool {
	switch c.Type {
	case ConditionTypeKeyword:
		return len(c.Keywords) > 0
	case ConditionTypeThreshold:
		return c.Threshold > 0 && c.WindowMin > 0
	case ConditionTypeWhitelist:
		return len(c.Whitelist) > 0
	}
	return false
}
