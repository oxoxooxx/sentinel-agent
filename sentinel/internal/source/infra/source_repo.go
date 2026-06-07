// Package infra — source 基礎設施層：來源 DB 存取
package infra

import (
	"context"
	"fmt"

	eventinfra "github.com/oxoxooxx/sentinel/internal/event/infra"
	"github.com/oxoxooxx/sentinel/internal/source/domain"
)

// SourceRepository 事件來源資料存取物件
type SourceRepository struct {
	db eventinfra.DB
}

// NewSourceRepository 建立 SourceRepository
func NewSourceRepository(db eventinfra.DB) *SourceRepository {
	return &SourceRepository{db: db}
}

// Save 新增（ID=0）或更新事件來源
func (r *SourceRepository) Save(ctx context.Context, src domain.Source) (domain.Source, error) {
	dbSrc := eventinfra.Source{
		ID:          src.ID,
		Name:        src.Name,
		Type:        string(src.Type),
		Address:     src.Address,
		Description: src.Description,
		CreatedAt:   src.CreatedAt,
	}

	saved, err := r.db.SaveSource(ctx, dbSrc)
	if err != nil {
		return domain.Source{}, fmt.Errorf("儲存來源失敗: %w", err)
	}

	src.ID = saved.ID
	return src, nil
}

// Delete 刪除事件來源
func (r *SourceRepository) Delete(ctx context.Context, id int64) error {
	if err := r.db.DeleteSource(ctx, id); err != nil {
		return fmt.Errorf("刪除來源失敗 id=%d: %w", id, err)
	}
	return nil
}
