// Package domain — event 領域層：Severity value object
package domain

// Severity 事件嚴重等級（value object）
type Severity string

const (
	SeverityCritical Severity = "critical" // 緊急，需立即處理
	SeverityHigh     Severity = "high"     // 高風險
	SeverityMedium   Severity = "medium"   // 中風險
	SeverityLow      Severity = "low"      // 低風險
	SeverityInfo     Severity = "info"     // 資訊，無需告警
)

// IsValid 檢查嚴重等級是否合法
func (s Severity) IsValid() bool {
	switch s {
	case SeverityCritical, SeverityHigh, SeverityMedium, SeverityLow, SeverityInfo:
		return true
	}
	return false
}

// IsAlertable 判斷此等級是否需要發送告警（medium 以上）
func (s Severity) IsAlertable() bool {
	return s == SeverityCritical || s == SeverityHigh || s == SeverityMedium
}

// String 回傳字串表示
func (s Severity) String() string {
	return string(s)
}

// ParseSeverity 將字串解析為 Severity，無法識別時回傳 SeverityInfo
func ParseSeverity(s string) Severity {
	sv := Severity(s)
	if sv.IsValid() {
		return sv
	}
	return SeverityInfo
}
