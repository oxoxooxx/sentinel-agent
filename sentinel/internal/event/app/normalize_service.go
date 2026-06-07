// Package app — event 應用層：消費 RawEventReceived，正規化後存 DB
package app

import (
	"context"
	"fmt"
	"log/slog"

	eventdomain "github.com/oxoxooxx/sentinel/internal/event/domain"
	eventinfra "github.com/oxoxooxx/sentinel/internal/event/infra"
	ingestiondomain "github.com/oxoxooxx/sentinel/internal/ingestion/domain"
	"github.com/oxoxooxx/sentinel/shared/eventbus"
)

// Decoder 是 ACL 解碼器介面，將原始事件轉為 NormalizedEvent
type Decoder interface {
	// CanDecode 判斷此解碼器是否能處理此來源的原始事件
	CanDecode(source, protocol string) bool
	// Decode 將原始 log 解碼為 NormalizedEvent
	Decode(sourceID int64, rawLog string) (eventdomain.NormalizedEvent, error)
}

// NormalizeService 消費 RawEventReceived，透過解碼器正規化後存入 DB
// 同時發布 EventNormalized domain event
type NormalizeService struct {
	db       eventinfra.DB
	decoders []Decoder
	bus      eventbus.Bus
}

// NewNormalizeService 建立 NormalizeService
func NewNormalizeService(db eventinfra.DB, bus eventbus.Bus, decoders ...Decoder) *NormalizeService {
	return &NormalizeService{
		db:       db,
		decoders: decoders,
		bus:      bus,
	}
}

// HandleRawEventReceived 處理 RawEventReceived domain event
func (s *NormalizeService) HandleRawEventReceived(ctx context.Context, evt ingestiondomain.RawEventReceived) error {
	raw := evt.RawEvent

	// 找到合適的解碼器
	var decoder Decoder
	for _, d := range s.decoders {
		if d.CanDecode(raw.Source, raw.Protocol) {
			decoder = d
			break
		}
	}
	if decoder == nil {
		return fmt.Errorf("找不到適合的解碼器 source=%s protocol=%s", raw.Source, raw.Protocol)
	}

	// 解碼正規化
	normalized, err := decoder.Decode(0, raw.RawData)
	if err != nil {
		return fmt.Errorf("解碼失敗: %w", err)
	}

	// 存入 DB（透過 infra 介面）
	dbEvent := eventinfra.Event{
		SourceID:  normalized.SourceID,
		RawLog:    normalized.RawLog,
		ParsedAt:  normalized.NormalAt,
		Severity:  eventinfra.Severity(normalized.Severity),
		Category:  normalized.Category,
		Message:   normalized.Message,
		ExtraJSON: normalized.ExtraJSON,
	}

	saved, err := s.db.SaveEvent(ctx, dbEvent)
	if err != nil {
		return fmt.Errorf("儲存正規化事件失敗: %w", err)
	}

	normalized.ID = saved.ID
	slog.Info("事件正規化完成", "event_id", saved.ID, "severity", normalized.Severity)

	// 發布 EventNormalized domain event
	outEvt := eventdomain.EventNormalized{Event: normalized}
	if err := s.bus.Publish(ctx, outEvt); err != nil {
		// 發布失敗不應影響儲存，僅記錄錯誤
		slog.Error("發布 EventNormalized 失敗", "event_id", saved.ID, "err", err)
	}

	return nil
}
