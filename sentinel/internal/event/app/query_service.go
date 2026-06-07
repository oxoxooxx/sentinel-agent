// Package app — event 應用層：Dashboard 查詢服務
package app

import (
	"context"
	"fmt"

	eventinfra "github.com/oxoxooxx/sentinel/internal/event/infra"
)

// QueryService 提供事件查詢功能，供 Dashboard 使用
type QueryService struct {
	db eventinfra.DB
}

// NewQueryService 建立 QueryService
func NewQueryService(db eventinfra.DB) *QueryService {
	return &QueryService{db: db}
}

// ListEvents 依篩選條件查詢事件列表
func (s *QueryService) ListEvents(ctx context.Context, filter eventinfra.EventFilter) ([]eventinfra.Event, int64, error) {
	total, err := s.db.CountEvents(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("統計事件數量失敗: %w", err)
	}

	events, err := s.db.QueryEvents(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("查詢事件失敗: %w", err)
	}

	return events, total, nil
}
