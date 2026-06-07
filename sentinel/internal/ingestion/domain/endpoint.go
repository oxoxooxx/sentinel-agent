// Package domain — ingestion 領域層：IngestionEndpoint aggregate
package domain

import "time"

// EndpointType 接收端點類型
type EndpointType string

const (
	EndpointTypeSyslog  EndpointType = "syslog"  // UDP syslog 接收
	EndpointTypeWebhook EndpointType = "webhook" // HTTP webhook 接收
)

// IngestionEndpoint 代表一個事件接收端點（aggregate root）
// 管理端點的生命週期與設定
type IngestionEndpoint struct {
	ID        string       // 唯一識別碼
	Type      EndpointType // 端點類型
	Address   string       // 監聽位址（如 ":514"、"/webhook"）
	Enabled   bool         // 是否啟用
	CreatedAt time.Time
}

// NewSyslogEndpoint 建立 UDP syslog 端點
func NewSyslogEndpoint(port int) *IngestionEndpoint {
	return &IngestionEndpoint{
		ID:        "syslog-default",
		Type:      EndpointTypeSyslog,
		Address:   formatPort(port),
		Enabled:   true,
		CreatedAt: time.Now(),
	}
}

// NewWebhookEndpoint 建立 HTTP webhook 端點
func NewWebhookEndpoint(path string) *IngestionEndpoint {
	return &IngestionEndpoint{
		ID:        "webhook-default",
		Type:      EndpointTypeWebhook,
		Address:   path,
		Enabled:   true,
		CreatedAt: time.Now(),
	}
}

// Enable 啟用端點，回傳新物件（immutable pattern）
func (e IngestionEndpoint) Enable() IngestionEndpoint {
	updated := e
	updated.Enabled = true
	return updated
}

// Disable 停用端點，回傳新物件（immutable pattern）
func (e IngestionEndpoint) Disable() IngestionEndpoint {
	updated := e
	updated.Enabled = false
	return updated
}

// formatPort 將 port 轉為監聽位址字串
func formatPort(port int) string {
	return ":" + itoa(port)
}

// itoa 簡易整數轉字串（避免引入 strconv）
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 10)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	return string(buf)
}
