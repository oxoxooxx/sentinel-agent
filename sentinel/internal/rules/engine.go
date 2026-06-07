// Package rules 實作告警規則引擎
package rules

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/oxoxooxx/sentinel/internal/storage"
)

// ConditionType 規則條件類型
type ConditionType string

const (
	CondKeyword   ConditionType = "keyword"   // 關鍵字比對
	CondThreshold ConditionType = "threshold" // 閾值觸發
	CondWhitelist ConditionType = "whitelist" // 白名單排除
)

// Condition 規則條件定義（對應 rules.cond_json）
type Condition struct {
	Type      ConditionType `json:"type"`
	Field     string        `json:"field"`              // 比對的欄位（"message", "raw_log" 等）
	Keywords  []string      `json:"keywords,omitempty"` // type=keyword 時使用
	Threshold int           `json:"threshold,omitempty"` // type=threshold 時使用（N 分鐘內 N 筆）
	WindowMin int           `json:"window_min,omitempty"` // 閾值時間窗口（分鐘）
	Whitelist []string      `json:"whitelist,omitempty"` // type=whitelist 時使用
	Severity  string        `json:"severity,omitempty"` // 觸發告警的嚴重等級
}

// MatchResult 規則比對結果
type MatchResult struct {
	Matched bool
	RuleID  int64
	Message string // 人類可讀的觸發說明
}

// Engine 規則引擎，負責載入規則並對事件進行比對
type Engine struct {
	db storage.DB
}

// NewEngine 建立規則引擎
func NewEngine(db storage.DB) *Engine {
	return &Engine{db: db}
}

// Evaluate 對單筆事件套用所有啟用中的規則，回傳符合的規則列表
func (e *Engine) Evaluate(ctx context.Context, event storage.Event) ([]MatchResult, error) {
	rules, err := e.db.ListRules(ctx, true) // 只取啟用的規則
	if err != nil {
		return nil, fmt.Errorf("載入規則失敗: %w", err)
	}

	var results []MatchResult
	for _, rule := range rules {
		result, err := e.matchRule(rule, event)
		if err != nil {
			// 個別規則錯誤不中斷整體評估，只記錄 log
			slog.Warn("規則評估失敗", "rule_id", rule.ID, "rule", rule.Name, "err", err)
			continue
		}
		if result.Matched {
			results = append(results, result)
		}
	}
	return results, nil
}

// matchRule 對單一規則進行比對
func (e *Engine) matchRule(rule storage.Rule, event storage.Event) (MatchResult, error) {
	var cond Condition
	if err := json.Unmarshal([]byte(rule.CondJSON), &cond); err != nil {
		return MatchResult{}, fmt.Errorf("解析規則 %d 條件失敗: %w", rule.ID, err)
	}

	// 取得要比對的欄位值
	fieldValue := extractField(event, cond.Field)

	switch cond.Type {
	case CondKeyword:
		return matchKeyword(rule, cond, fieldValue), nil

	case CondWhitelist:
		return matchWhitelist(rule, cond, fieldValue), nil

	case CondThreshold:
		// 閾值比對需要查詢歷史資料，此處標記為待實作
		// TODO: 實作滑動視窗計數（需要 CountEvents）
		slog.Debug("閾值規則暫未實作", "rule_id", rule.ID)
		return MatchResult{Matched: false, RuleID: rule.ID}, nil

	default:
		return MatchResult{}, fmt.Errorf("未知的規則類型: %s", cond.Type)
	}
}

// matchKeyword 關鍵字比對（含任一關鍵字即觸發）
func matchKeyword(rule storage.Rule, cond Condition, fieldValue string) MatchResult {
	lowerValue := strings.ToLower(fieldValue)
	for _, kw := range cond.Keywords {
		if strings.Contains(lowerValue, strings.ToLower(kw)) {
			return MatchResult{
				Matched: true,
				RuleID:  rule.ID,
				Message: fmt.Sprintf("規則「%s」觸發：偵測到關鍵字 %q", rule.Name, kw),
			}
		}
	}
	return MatchResult{Matched: false, RuleID: rule.ID}
}

// matchWhitelist 白名單比對（任一白名單項目命中則排除告警）
func matchWhitelist(rule storage.Rule, cond Condition, fieldValue string) MatchResult {
	lowerValue := strings.ToLower(fieldValue)
	for _, item := range cond.Whitelist {
		if strings.Contains(lowerValue, strings.ToLower(item)) {
			// 在白名單內，不觸發告警
			return MatchResult{Matched: false, RuleID: rule.ID}
		}
	}
	// 不在白名單內，觸發告警
	return MatchResult{
		Matched: true,
		RuleID:  rule.ID,
		Message: fmt.Sprintf("規則「%s」觸發：來源不在白名單內", rule.Name),
	}
}

// extractField 從 Event 中取出指定欄位的字串值
func extractField(event storage.Event, field string) string {
	switch field {
	case "message":
		return event.Message
	case "raw_log":
		return event.RawLog
	case "category":
		return event.Category
	case "severity":
		return string(event.Severity)
	default:
		// 找不到欄位時，預設比對 message
		return event.Message
	}
}
