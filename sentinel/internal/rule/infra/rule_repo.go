// Package infra — rule 基礎設施層：規則 DB 存取
package infra

import (
	"context"
	"fmt"

	eventinfra "github.com/oxoxooxx/sentinel/internal/event/infra"
	"github.com/oxoxooxx/sentinel/internal/rule/domain"
)

// RuleRepository 規則資料存取物件
// 目前透過 eventinfra.DB 統一存取，後續可獨立出專屬 DB 介面
type RuleRepository struct {
	db eventinfra.DB
}

// NewRuleRepository 建立 RuleRepository
func NewRuleRepository(db eventinfra.DB) *RuleRepository {
	return &RuleRepository{db: db}
}

// ListEnabled 列出所有啟用中的規則
func (r *RuleRepository) ListEnabled(ctx context.Context) ([]domain.Rule, error) {
	dbRules, err := r.db.ListRules(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("查詢規則失敗: %w", err)
	}

	rules := make([]domain.Rule, 0, len(dbRules))
	for _, dbRule := range dbRules {
		rule, err := toDomainRule(dbRule)
		if err != nil {
			// 個別規則格式錯誤不中斷整體，跳過
			continue
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

// Save 新增（ID=0）或更新規則
func (r *RuleRepository) Save(ctx context.Context, rule domain.Rule) (domain.Rule, error) {
	condJSON, err := rule.CondJSON()
	if err != nil {
		return domain.Rule{}, fmt.Errorf("序列化規則條件失敗: %w", err)
	}

	dbRule := eventinfra.Rule{
		ID:          rule.ID,
		Name:        rule.Name,
		Description: rule.Description,
		Enabled:     rule.Enabled,
		CondJSON:    condJSON,
		CreatedAt:   rule.CreatedAt,
	}

	saved, err := r.db.SaveRule(ctx, dbRule)
	if err != nil {
		return domain.Rule{}, fmt.Errorf("儲存規則失敗: %w", err)
	}

	rule.ID = saved.ID
	return rule, nil
}

// toDomainRule 將 DB 規則模型轉換為 domain.Rule
func toDomainRule(dbRule eventinfra.Rule) (domain.Rule, error) {
	return domain.Rule{
		ID:          dbRule.ID,
		Name:        dbRule.Name,
		Description: dbRule.Description,
		Enabled:     dbRule.Enabled,
		CreatedAt:   dbRule.CreatedAt,
	}, nil
}
