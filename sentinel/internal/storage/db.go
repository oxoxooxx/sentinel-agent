// Package storage 定義儲存後端的統一介面與資料模型
package storage

import (
	"context"
	"time"
)

// Severity 事件嚴重等級
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// AlertStatus 告警狀態
type AlertStatus string

const (
	AlertStatusPending    AlertStatus = "pending"    // 尚未處理
	AlertStatusSent       AlertStatus = "sent"       // 已送出通知
	AlertStatusAcked      AlertStatus = "acked"      // 已確認
	AlertStatusResolved   AlertStatus = "resolved"   // 已解決
)

// Event 代表一筆從外部來源接收到的安全事件
type Event struct {
	ID        int64     `json:"id"`
	SourceID  int64     `json:"source_id"`  // 對應 sources 表
	RawLog    string    `json:"raw_log"`    // 原始 log 內容
	ParsedAt  time.Time `json:"parsed_at"`  // 解析時間
	Severity  Severity  `json:"severity"`
	Category  string    `json:"category"`   // 如 "firewall", "nas", "webhook"
	Message   string    `json:"message"`    // 人類可讀的摘要
	ExtraJSON string    `json:"extra_json"` // 額外結構化資料（JSON 字串）
}

// EventFilter 查詢事件的篩選條件
type EventFilter struct {
	SourceID *int64    // 指定來源
	Severity *Severity // 指定嚴重等級
	From     *time.Time
	To       *time.Time
	Limit    int // 0 表示不限制
	Offset   int
}

// Alert 代表一筆告警記錄
type Alert struct {
	ID        int64       `json:"id"`
	EventID   int64       `json:"event_id"`   // 觸發此告警的事件
	RuleID    int64       `json:"rule_id"`    // 觸發此告警的規則
	Status    AlertStatus `json:"status"`
	Channel   string      `json:"channel"`    // 通知頻道，如 "line", "telegram"
	Message   string      `json:"message"`    // 送出的告警訊息
	CreatedAt time.Time   `json:"created_at"`
	SentAt    *time.Time  `json:"sent_at,omitempty"`
	AckedAt   *time.Time  `json:"acked_at,omitempty"`
}

// Rule 代表一條告警規則
type Rule struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Enabled     bool      `json:"enabled"`
	CondJSON    string    `json:"cond_json"` // 規則條件（JSON 字串）
	CreatedAt   time.Time `json:"created_at"`
}

// Source 代表一個事件來源（如某台 FortiGate、某個 NAS）
type Source struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`        // "fortigate", "nas", "webhook"
	Address     string    `json:"address"`     // IP 或 URL
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// NotificationChannel 代表一個通知頻道設定
type NotificationChannel struct {
	ID         int64     `json:"id"`
	Type       string    `json:"type"`        // "line", "telegram", "teams"
	Name       string    `json:"name"`
	ConfigJSON string    `json:"config_json"` // 頻道設定（JSON 字串，含 token 等）
	Enabled    bool      `json:"enabled"`
	CreatedAt  time.Time `json:"created_at"`
}

// DB 是儲存後端的統一介面，支援 SQLite 與 PostgreSQL 切換
type DB interface {
	// --- 初始化 ---

	// Migrate 執行資料庫 schema 建立／升級
	Migrate(ctx context.Context) error

	// Close 釋放資料庫連線
	Close() error

	// --- 事件 ---

	// SaveEvent 儲存一筆新事件，回傳帶有 ID 的 Event
	SaveEvent(ctx context.Context, e Event) (Event, error)

	// QueryEvents 依篩選條件查詢事件列表
	QueryEvents(ctx context.Context, filter EventFilter) ([]Event, error)

	// CountEvents 回傳符合條件的事件數量
	CountEvents(ctx context.Context, filter EventFilter) (int64, error)

	// --- 告警 ---

	// SaveAlert 儲存一筆新告警，回傳帶有 ID 的 Alert
	SaveAlert(ctx context.Context, a Alert) (Alert, error)

	// UpdateAlertStatus 更新告警狀態
	UpdateAlertStatus(ctx context.Context, id int64, status AlertStatus) error

	// QueryAlerts 查詢告警列表
	QueryAlerts(ctx context.Context, status *AlertStatus, limit, offset int) ([]Alert, error)

	// --- 規則 ---

	// SaveRule 新增或更新規則（ID=0 表示新增）
	SaveRule(ctx context.Context, r Rule) (Rule, error)

	// DeleteRule 刪除規則
	DeleteRule(ctx context.Context, id int64) error

	// ListRules 列出所有規則（可依 enabled 篩選）
	ListRules(ctx context.Context, enabledOnly bool) ([]Rule, error)

	// --- 來源 ---

	// SaveSource 新增或更新事件來源（ID=0 表示新增）
	SaveSource(ctx context.Context, s Source) (Source, error)

	// ListSources 列出所有事件來源
	ListSources(ctx context.Context) ([]Source, error)

	// DeleteSource 刪除事件來源
	DeleteSource(ctx context.Context, id int64) error

	// --- 通知頻道 ---

	// ListNotificationChannels 列出所有通知頻道
	ListNotificationChannels(ctx context.Context) ([]NotificationChannel, error)
}
