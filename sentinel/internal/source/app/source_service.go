// Package app — source 應用層：來源 CRUD
package app

import (
	"context"
	"fmt"
	"log/slog"

	eventinfra "github.com/oxoxooxx/sentinel/internal/event/infra"
	"github.com/oxoxooxx/sentinel/internal/source/domain"
	"github.com/oxoxooxx/sentinel/shared/eventbus"
)

// SourceService 提供事件來源的 CRUD 操作
type SourceService struct {
	db  eventinfra.DB
	bus eventbus.Bus
}

// NewSourceService 建立 SourceService
func NewSourceService(db eventinfra.DB, bus eventbus.Bus) *SourceService {
	return &SourceService{db: db, bus: bus}
}

// Create 新增一個事件來源
func (s *SourceService) Create(ctx context.Context, name string, srcType domain.SourceType, address, description string) (domain.Source, error) {
	src, err := domain.NewSource(name, srcType, address, description)
	if err != nil {
		return domain.Source{}, fmt.Errorf("建立來源失敗: %w", err)
	}

	dbSrc := eventinfra.Source{
		Name:        src.Name,
		Type:        string(src.Type),
		Address:     src.Address,
		Description: src.Description,
		CreatedAt:   src.CreatedAt,
	}

	saved, err := s.db.SaveSource(ctx, dbSrc)
	if err != nil {
		return domain.Source{}, fmt.Errorf("儲存來源失敗: %w", err)
	}

	src.ID = saved.ID
	slog.Info("事件來源已建立", "id", src.ID, "name", src.Name, "type", src.Type)

	// 發布 SourceRegistered domain event
	evt := domain.SourceRegistered{Source: src}
	if err := s.bus.Publish(ctx, evt); err != nil {
		slog.Error("發布 SourceRegistered 失敗", "source_id", src.ID, "err", err)
	}

	return src, nil
}

// List 列出所有事件來源
func (s *SourceService) List(ctx context.Context) ([]eventinfra.Source, error) {
	sources, err := s.db.ListSources(ctx)
	if err != nil {
		return nil, fmt.Errorf("查詢來源失敗: %w", err)
	}
	return sources, nil
}

// Delete 刪除一個事件來源
func (s *SourceService) Delete(ctx context.Context, id int64) error {
	if err := s.db.DeleteSource(ctx, id); err != nil {
		return fmt.Errorf("刪除來源失敗 id=%d: %w", id, err)
	}
	slog.Info("事件來源已刪除", "id", id)
	return nil
}
