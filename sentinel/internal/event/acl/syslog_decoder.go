// Package acl — event Anti-Corruption Layer：通用 syslog 解析
package acl

import (
	"fmt"
	"strings"

	"github.com/oxoxooxx/sentinel/internal/event/domain"
)

// SyslogDecoder 通用 syslog 解析器（RFC 5424 / BSD syslog 簡易實作）
// 適用於非 FortiGate 的 syslog 來源
type SyslogDecoder struct{}

// NewSyslogDecoder 建立通用 syslog 解碼器
func NewSyslogDecoder() *SyslogDecoder {
	return &SyslogDecoder{}
}

// CanDecode 通用 syslog 解碼器接受所有 UDP 來源（作為 fallback）
func (d *SyslogDecoder) CanDecode(source, protocol string) bool {
	return protocol == "udp" || protocol == "tcp"
}

// Decode 將通用 syslog 原始 log 解碼為 NormalizedEvent
func (d *SyslogDecoder) Decode(sourceID int64, rawLog string) (domain.NormalizedEvent, error) {
	if rawLog == "" {
		return domain.NormalizedEvent{}, fmt.Errorf("rawLog 不可為空")
	}

	severity := parseSyslogPriority(rawLog)
	message := stripPriority(rawLog)

	evt := domain.NewNormalizedEvent(sourceID, rawLog, severity, "syslog", message)
	return evt, nil
}

// parseSyslogPriority 從 RFC 5424 的 <PRI> 欄位解析嚴重等級
// PRI = Facility * 8 + Severity，Severity 0-7（0=Emergency, 7=Debug）
func parseSyslogPriority(rawLog string) domain.Severity {
	if len(rawLog) > 0 && rawLog[0] == '<' {
		end := strings.IndexByte(rawLog, '>')
		if end > 1 {
			pri := atoi(rawLog[1:end])
			syslogSev := pri % 8
			switch {
			case syslogSev <= 2: // Emergency, Alert, Critical
				return domain.SeverityCritical
			case syslogSev == 3: // Error
				return domain.SeverityHigh
			case syslogSev == 4: // Warning
				return domain.SeverityMedium
			case syslogSev == 5: // Notice
				return domain.SeverityLow
			}
		}
	}
	return domain.SeverityInfo
}

// stripPriority 移除 syslog 訊息的 <PRI> 前綴
func stripPriority(rawLog string) string {
	if len(rawLog) > 0 && rawLog[0] == '<' {
		if end := strings.IndexByte(rawLog, '>'); end >= 0 {
			return strings.TrimSpace(rawLog[end+1:])
		}
	}
	return rawLog
}

// atoi 簡易字串轉整數（不依賴 strconv）
func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return n
}
