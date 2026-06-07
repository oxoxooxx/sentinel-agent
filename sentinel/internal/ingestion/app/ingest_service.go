// Package app — ingestion 應用層：監聽生命週期管理，發布 domain event
package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/oxoxooxx/sentinel/internal/ingestion/domain"
	"github.com/oxoxooxx/sentinel/shared/eventbus"
)

// IngestService 管理 ingestion 端點的生命週期
// 接收到原始事件後透過 eventbus 發布 RawEventReceived domain event
type IngestService struct {
	bus eventbus.Bus
}

// NewIngestService 建立 IngestService
func NewIngestService(bus eventbus.Bus) *IngestService {
	return &IngestService{bus: bus}
}

// HandleRawEvent 接收原始事件並發布 domain event
func (s *IngestService) HandleRawEvent(ctx context.Context, endpointID string, raw domain.RawEvent) error {
	if raw.IsEmpty() {
		return fmt.Errorf("原始事件資料不可為空")
	}

	evt := domain.RawEventReceived{
		EndpointID: endpointID,
		RawEvent:   raw,
	}

	if err := s.bus.Publish(ctx, evt); err != nil {
		return fmt.Errorf("發布 RawEventReceived 失敗: %w", err)
	}

	slog.Debug("已發布 RawEventReceived", "endpoint", endpointID, "source", raw.Source)
	return nil
}
