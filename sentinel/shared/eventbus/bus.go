// Package eventbus — in-process sync event bus 介面與實作
// 供各 bounded context 之間透過 domain event 進行解耦通訊
package eventbus

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

// Event 是所有 domain event 必須實作的介面
type Event interface {
	// EventName 回傳事件的唯一名稱（如 "ingestion.RawEventReceived"）
	EventName() string
}

// Handler 是事件處理函式的型別
type Handler func(ctx context.Context, event Event) error

// Bus 是 in-process event bus 的介面
type Bus interface {
	// Publish 發布一個 domain event，同步呼叫所有已訂閱的 handler
	Publish(ctx context.Context, event Event) error

	// Subscribe 訂閱指定事件名稱的 handler
	Subscribe(eventName string, handler Handler)
}

// SyncBus 是同步（sync）in-process event bus 實作
// 適合單體應用；若需要非同步或跨進程，可替換為其他實作
type SyncBus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

// NewSyncBus 建立 SyncBus 實例
func NewSyncBus() *SyncBus {
	return &SyncBus{
		handlers: make(map[string][]Handler),
	}
}

// Subscribe 訂閱指定事件名稱的 handler
func (b *SyncBus) Subscribe(eventName string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], handler)
	slog.Debug("eventbus: 訂閱事件", "event", eventName)
}

// Publish 發布 domain event，依序同步呼叫所有 handler
// 任一 handler 回傳錯誤時記錄 log，但不中斷其他 handler 的執行
func (b *SyncBus) Publish(ctx context.Context, event Event) error {
	b.mu.RLock()
	handlers := b.handlers[event.EventName()]
	b.mu.RUnlock()

	if len(handlers) == 0 {
		slog.Debug("eventbus: 沒有訂閱者", "event", event.EventName())
		return nil
	}

	var lastErr error
	for _, h := range handlers {
		if err := h(ctx, event); err != nil {
			slog.Error("eventbus: handler 執行失敗",
				"event", event.EventName(),
				"err", err,
			)
			lastErr = fmt.Errorf("handler 失敗 [%s]: %w", event.EventName(), err)
		}
	}
	return lastErr
}
