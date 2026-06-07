// Package domain — rule 領域層：Rule aggregate
package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

// RuleType 規則類型
type RuleType string

const (
	RuleTypeKeyword   RuleType = "keyword"   // 關鍵字比對
	RuleTypeThreshold RuleType = "threshold" // 閾值觸發
	RuleTypeWhitelist RuleType = "whitelist" // 白名單排除
)

// Rule 代表一條告警規則（aggregate root）
type Rule struct {
	ID          int64
	Name        string
	Description string
	Enabled     bool
	Type        RuleType
	Condition   Condition // 規則條件
	CreatedAt   time.Time
}

// NewKeywordRule 建立關鍵字比對規則
func NewKeywordRule(name, description, field string, keywords []string) (Rule, error) {
	if name == "" {
		return Rule{}, fmt.Errorf("規則名稱不可為空")
	}
	if len(keywords) == 0 {
		return Rule{}, fmt.Errorf("關鍵字列表不可為空")
	}

	return Rule{
		Name:        name,
		Description: description,
		Enabled:     true,
		Type:        RuleTypeKeyword,
		Condition: Condition{
			Type:     ConditionTypeKeyword,
			Field:    field,
			Keywords: keywords,
		},
		CreatedAt: time.Now(),
	}, nil
}

// CondJSON 序列化 Condition 為 JSON 字串（供 DB 儲存使用）
func (r Rule) CondJSON() (string, error) {
	b, err := json.Marshal(r.Condition)
	if err != nil {
		return "{}", fmt.Errorf("序列化規則條件失敗: %w", err)
	}
	return string(b), nil
}

// Enable 啟用規則，回傳新物件（immutable pattern）
func (r Rule) Enable() Rule {
	updated := r
	updated.Enabled = true
	return updated
}

// Disable 停用規則，回傳新物件（immutable pattern）
func (r Rule) Disable() Rule {
	updated := r
	updated.Enabled = false
	return updated
}
