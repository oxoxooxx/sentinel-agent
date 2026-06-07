// Package domain — rule 領域層：domain events
package domain

// RuleMatched 是當規則比對成功時發出的 domain event
// 消費者（如 alert/app/dispatcher.go）訂閱此事件觸發告警派送
type RuleMatched struct {
	RuleID   int64  // 觸發的規則 ID
	RuleName string // 規則名稱（方便 log）
	EventID  int64  // 觸發此規則的事件 ID
	Message  string // 人類可讀的觸發說明
}

// EventName 實作 domain event 識別介面
func (e RuleMatched) EventName() string {
	return "rule.RuleMatched"
}
