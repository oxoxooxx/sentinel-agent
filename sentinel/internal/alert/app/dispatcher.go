// Package app — alert 應用層：告警派送，將告警透過各通知頻道送出
package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/oxoxooxx/sentinel/config"
	eventinfra "github.com/oxoxooxx/sentinel/internal/event/infra"
	alertinfra "github.com/oxoxooxx/sentinel/internal/alert/infra"
)

// Notifier 是通知頻道的統一介面
type Notifier interface {
	// Send 發送告警訊息，回傳錯誤
	Send(ctx context.Context, msg string) error
	// Name 回傳頻道名稱（用於 log 與記錄）
	Name() string
}

// Dispatcher 負責管理多個通知頻道，並派送告警
type Dispatcher struct {
	db        eventinfra.DB
	notifiers []Notifier
}

// NewDispatcher 根據設定建立 Dispatcher
// 依序初始化 LINE、Telegram、Teams 頻道
func NewDispatcher(db eventinfra.DB, cfg config.AlertConfig) (*Dispatcher, error) {
	d := &Dispatcher{db: db}

	for _, ch := range cfg.Channels {
		var n Notifier
		var err error

		switch ch.Type {
		case "line":
			n, err = alertinfra.NewLINENotifier(ch.Token)
		case "telegram":
			n, err = alertinfra.NewTelegramNotifier(ch.BotToken, ch.ChatID)
		case "teams":
			n, err = alertinfra.NewTeamsNotifier(ch.WebhookURL)
		default:
			slog.Warn("未知的通知頻道類型，略過", "type", ch.Type)
			continue
		}

		if err != nil {
			return nil, fmt.Errorf("初始化 %s 通知頻道失敗: %w", ch.Type, err)
		}
		d.notifiers = append(d.notifiers, n)
		slog.Info("通知頻道就緒", "type", ch.Type)
	}

	return d, nil
}

// Dispatch 將事件觸發的告警送往所有已設定的通知頻道
// 同時將告警記錄寫入 DB
func (d *Dispatcher) Dispatch(ctx context.Context, event eventinfra.Event, ruleID int64, message string) error {
	if len(d.notifiers) == 0 {
		slog.Warn("沒有可用的通知頻道，告警未送出", "event_id", event.ID)
		return nil
	}

	// 組合告警訊息
	alertMsg := formatAlertMessage(event, message)

	var lastErr error
	for _, n := range d.notifiers {
		// 先建立 pending 記錄
		a := eventinfra.Alert{
			EventID: event.ID,
			RuleID:  ruleID,
			Status:  eventinfra.AlertStatusPending,
			Channel: n.Name(),
			Message: alertMsg,
		}

		saved, err := d.db.SaveAlert(ctx, a)
		if err != nil {
			slog.Error("儲存告警記錄失敗", "channel", n.Name(), "err", err)
			lastErr = err
			continue
		}

		// 送出通知
		if err := n.Send(ctx, alertMsg); err != nil {
			slog.Error("告警送出失敗", "channel", n.Name(), "err", err)
			_ = d.db.UpdateAlertStatus(ctx, saved.ID, eventinfra.AlertStatusPending) // 維持 pending 以便重試
			lastErr = err
			continue
		}

		// 更新為已送出
		if err := d.db.UpdateAlertStatus(ctx, saved.ID, eventinfra.AlertStatusSent); err != nil {
			slog.Warn("更新告警狀態失敗", "alert_id", saved.ID, "err", err)
		}

		slog.Info("告警已送出", "channel", n.Name(), "event_id", event.ID)
	}

	return lastErr
}

// formatAlertMessage 將事件資訊格式化為告警訊息
func formatAlertMessage(event eventinfra.Event, ruleMsg string) string {
	severityEmoji := map[eventinfra.Severity]string{
		eventinfra.SeverityCritical: "🔴",
		eventinfra.SeverityWarning:  "🟡",
		eventinfra.SeverityInfo:     "🔵",
	}

	emoji := severityEmoji[event.Severity]
	if emoji == "" {
		emoji = "⚠️"
	}

	return fmt.Sprintf(
		"%s [Sentinel 告警]\n嚴重等級: %s\n分類: %s\n訊息: %s\n觸發規則: %s",
		emoji,
		string(event.Severity),
		event.Category,
		event.Message,
		ruleMsg,
	)
}
