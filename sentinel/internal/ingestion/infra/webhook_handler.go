// Package infra — ingestion 基礎設施層：HTTP webhook 接收器
package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	eventinfra "github.com/oxoxooxx/sentinel/internal/event/infra"
)

// WebhookPayload 是外部系統 POST 過來的 JSON 格式
type WebhookPayload struct {
	Source   string          `json:"source"`   // 來源識別字串，如 "synology-nas"
	Severity string          `json:"severity"` // "info" | "warning" | "critical"
	Message  string          `json:"message"`  // 人類可讀摘要
	RawLog   string          `json:"raw_log"`  // 原始 log（選填）
	Extra    json.RawMessage `json:"extra"`    // 自訂擴充欄位
}

// WebhookServer 提供 HTTP webhook 接收端點
type WebhookServer struct {
	db     eventinfra.DB
	secret string // Bearer token 驗證（空字串表示不驗證）
}

// NewWebhookServer 建立 webhook 接收器
func NewWebhookServer(db eventinfra.DB, secret string) *WebhookServer {
	return &WebhookServer{db: db, secret: secret}
}

// Handler 回傳 http.Handler，掛載到 HTTP 路由
// 端點：POST /webhook
func (w *WebhookServer) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /webhook", w.handleWebhook)
	return mux
}

// handleWebhook 處理 webhook POST 請求
func (w *WebhookServer) handleWebhook(rw http.ResponseWriter, r *http.Request) {
	// Bearer token 驗證（若有設定）
	if w.secret != "" {
		token := r.Header.Get("Authorization")
		expected := "Bearer " + w.secret
		if token != expected {
			http.Error(rw, "unauthorized", http.StatusUnauthorized)
			return
		}
	}

	// 限制 body 大小（最大 1MB）
	r.Body = http.MaxBytesReader(rw, r.Body, 1<<20)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(rw, "讀取請求失敗", http.StatusBadRequest)
		return
	}

	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(rw, fmt.Sprintf("JSON 解析失敗: %v", err), http.StatusBadRequest)
		return
	}

	if payload.Message == "" {
		http.Error(rw, "message 欄位不可為空", http.StatusBadRequest)
		return
	}

	// 將 webhook 事件存入 DB
	ctx := r.Context()
	if err := w.save(ctx, payload, r.RemoteAddr, string(body)); err != nil {
		slog.Error("webhook 事件儲存失敗", "err", err)
		http.Error(rw, "儲存失敗，請稍後重試", http.StatusInternalServerError)
		return
	}

	slog.Info("收到 webhook 事件", "source", payload.Source, "severity", payload.Severity)
	rw.WriteHeader(http.StatusAccepted)
	rw.Write([]byte(`{"ok":true}`))
}

// save 將 webhook payload 轉換為 Event 並儲存
func (w *WebhookServer) save(ctx context.Context, payload WebhookPayload, remoteAddr, rawBody string) error {
	sev := parseWebhookSeverity(payload.Severity)

	rawLog := payload.RawLog
	if rawLog == "" {
		rawLog = rawBody
	}

	extra := "{}"
	if len(payload.Extra) > 0 {
		extra = string(payload.Extra)
	}

	event := eventinfra.Event{
		SourceID:  0, // webhook 來源不與特定 source 綁定，後續可依 payload.Source 查詢
		RawLog:    rawLog,
		ParsedAt:  time.Now(),
		Severity:  sev,
		Category:  "webhook",
		Message:   payload.Message,
		ExtraJSON: fmt.Sprintf(`{"source": "%s", "remote_addr": "%s", "extra": %s}`, payload.Source, remoteAddr, extra),
	}

	if _, err := w.db.SaveEvent(ctx, event); err != nil {
		return fmt.Errorf("SaveEvent 失敗: %w", err)
	}
	return nil
}

// parseWebhookSeverity 將字串轉換為 Severity
func parseWebhookSeverity(s string) eventinfra.Severity {
	switch s {
	case "critical":
		return eventinfra.SeverityCritical
	case "warning":
		return eventinfra.SeverityWarning
	default:
		return eventinfra.SeverityInfo
	}
}
