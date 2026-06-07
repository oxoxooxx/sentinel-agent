// Package acl — event Anti-Corruption Layer：FortiGate syslog 格式解析
package acl

import (
	"fmt"
	"strings"

	"github.com/oxoxooxx/sentinel/internal/event/domain"
)

// FortigateDecoder 解析 FortiGate syslog 格式（key=value 配對）
type FortigateDecoder struct{}

// NewFortigateDecoder 建立 FortiGate 解碼器
func NewFortigateDecoder() *FortigateDecoder {
	return &FortigateDecoder{}
}

// CanDecode 判斷是否為 FortiGate syslog（UDP 協議且包含 FortiGate 特徵）
func (d *FortigateDecoder) CanDecode(source, protocol string) bool {
	return protocol == "udp"
}

// Decode 將 FortiGate syslog 原始 log 解碼為 NormalizedEvent
func (d *FortigateDecoder) Decode(sourceID int64, rawLog string) (domain.NormalizedEvent, error) {
	if rawLog == "" {
		return domain.NormalizedEvent{}, fmt.Errorf("rawLog 不可為空")
	}

	// 解析 key=value 格式（FortiGate syslog 標準格式）
	fields := parseKV(rawLog)

	severity := parseFGSeverity(fields["level"])
	category := fields["type"]
	if category == "" {
		category = "syslog"
	}

	message := fields["msg"]
	if message == "" {
		message = rawLog
	}

	extra := buildExtraJSON(fields)

	evt := domain.NewNormalizedEvent(sourceID, rawLog, severity, category, message).
		WithExtra(extra)

	return evt, nil
}

// parseKV 將 FortiGate key=value 格式字串解析為 map
func parseKV(s string) map[string]string {
	result := make(map[string]string)
	// 簡易解析：以空格分隔，每個 token 含 "="
	parts := strings.Fields(s)
	for _, part := range parts {
		if idx := strings.IndexByte(part, '='); idx > 0 {
			key := part[:idx]
			val := strings.Trim(part[idx+1:], `"`)
			result[key] = val
		}
	}
	return result
}

// parseFGSeverity 將 FortiGate level 欄位轉換為 Severity
func parseFGSeverity(level string) domain.Severity {
	switch strings.ToLower(level) {
	case "emergency", "alert", "critical":
		return domain.SeverityCritical
	case "error":
		return domain.SeverityHigh
	case "warning":
		return domain.SeverityMedium
	case "notice", "information":
		return domain.SeverityLow
	default:
		return domain.SeverityInfo
	}
}

// buildExtraJSON 將解析後的 fields 轉為 JSON 字串（簡易實作）
func buildExtraJSON(fields map[string]string) string {
	if len(fields) == 0 {
		return "{}"
	}
	var parts []string
	for k, v := range fields {
		// 安全地轉義雙引號
		v = strings.ReplaceAll(v, `"`, `\"`)
		parts = append(parts, fmt.Sprintf(`"%s":"%s"`, k, v))
	}
	return "{" + strings.Join(parts, ",") + "}"
}
