// Package acl — event Anti-Corruption Layer：webhook JSON 解析
package acl

import (
	"encoding/json"
	"fmt"

	"github.com/oxoxooxx/sentinel/internal/event/domain"
)

// webhookPayload 是外部系統 POST 過來的 JSON 格式（對應 ingestion/infra/webhook_handler.go）
type webhookPayload struct {
	Source   string `json:"source"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	RawLog   string `json:"raw_log"`
}

// WebhookDecoder 解析 webhook JSON 格式
type WebhookDecoder struct{}

// NewWebhookDecoder 建立 webhook 解碼器
func NewWebhookDecoder() *WebhookDecoder {
	return &WebhookDecoder{}
}

// CanDecode 判斷是否為 HTTP webhook 來源
func (d *WebhookDecoder) CanDecode(source, protocol string) bool {
	return protocol == "http"
}

// Decode 將 webhook JSON 原始 log 解碼為 NormalizedEvent
func (d *WebhookDecoder) Decode(sourceID int64, rawLog string) (domain.NormalizedEvent, error) {
	if rawLog == "" {
		return domain.NormalizedEvent{}, fmt.Errorf("rawLog 不可為空")
	}

	var payload webhookPayload
	if err := json.Unmarshal([]byte(rawLog), &payload); err != nil {
		// JSON 解析失敗時，以原始內容作為訊息
		return domain.NewNormalizedEvent(sourceID, rawLog, domain.SeverityInfo, "webhook", rawLog), nil
	}

	severity := parseWebhookSev(payload.Severity)
	message := payload.Message
	if message == "" {
		message = rawLog
	}

	raw := payload.RawLog
	if raw == "" {
		raw = rawLog
	}

	extra := fmt.Sprintf(`{"source":"%s"}`, payload.Source)
	evt := domain.NewNormalizedEvent(sourceID, raw, severity, "webhook", message).
		WithExtra(extra)

	return evt, nil
}

// parseWebhookSev 將字串嚴重等級解析為 domain.Severity
func parseWebhookSev(s string) domain.Severity {
	switch s {
	case "critical":
		return domain.SeverityCritical
	case "high":
		return domain.SeverityHigh
	case "medium":
		return domain.SeverityMedium
	case "low":
		return domain.SeverityLow
	default:
		return domain.SeverityInfo
	}
}
