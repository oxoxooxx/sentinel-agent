// Package storage — SQLite 實作
package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3" // 注入 SQLite 驅動
)

// SQLiteDB 實作 DB 介面，底層使用 SQLite
type SQLiteDB struct {
	db *sql.DB
}

// NewSQLite 開啟（或建立）指定路徑的 SQLite 資料庫
func NewSQLite(path string) (*SQLiteDB, error) {
	// WAL 模式提升並發讀取效能；foreign_keys 確保關聯完整性
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_foreign_keys=on", path)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("開啟 SQLite 失敗 %s: %w", path, err)
	}

	// SQLite 單一寫入者，最大連線數設 1 避免鎖定衝突
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	return &SQLiteDB{db: db}, nil
}

// Migrate 執行 schema 建立
func (s *SQLiteDB) Migrate(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, SQLiteSchema); err != nil {
		return fmt.Errorf("SQLite migrate 失敗: %w", err)
	}
	return nil
}

// Close 關閉資料庫連線
func (s *SQLiteDB) Close() error {
	return s.db.Close()
}

// ---- 事件 ----

// SaveEvent 將事件寫入 events 資料表
func (s *SQLiteDB) SaveEvent(ctx context.Context, e Event) (Event, error) {
	const q = `
		INSERT INTO events (source_id, raw_log, parsed_at, severity, category, message, extra_json)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	if e.ParsedAt.IsZero() {
		e.ParsedAt = time.Now()
	}
	if e.ExtraJSON == "" {
		e.ExtraJSON = "{}"
	}

	res, err := s.db.ExecContext(ctx, q,
		e.SourceID, e.RawLog, e.ParsedAt, string(e.Severity),
		e.Category, e.Message, e.ExtraJSON,
	)
	if err != nil {
		return Event{}, fmt.Errorf("SaveEvent 失敗: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return Event{}, fmt.Errorf("取得事件 ID 失敗: %w", err)
	}

	e.ID = id
	return e, nil
}

// QueryEvents 依篩選條件查詢事件
func (s *SQLiteDB) QueryEvents(ctx context.Context, filter EventFilter) ([]Event, error) {
	q, args := buildEventQuery(filter, false)
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("QueryEvents 失敗: %w", err)
	}
	defer rows.Close()

	return scanEvents(rows)
}

// CountEvents 回傳符合條件的事件數量
func (s *SQLiteDB) CountEvents(ctx context.Context, filter EventFilter) (int64, error) {
	q, args := buildEventQuery(filter, true)
	var count int64
	if err := s.db.QueryRowContext(ctx, q, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("CountEvents 失敗: %w", err)
	}
	return count, nil
}

// ---- 告警 ----

// SaveAlert 將告警寫入 alerts 資料表
func (s *SQLiteDB) SaveAlert(ctx context.Context, a Alert) (Alert, error) {
	const q = `
		INSERT INTO alerts (event_id, rule_id, status, channel, message, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now()
	}
	if a.Status == "" {
		a.Status = AlertStatusPending
	}

	res, err := s.db.ExecContext(ctx, q,
		a.EventID, a.RuleID, string(a.Status), a.Channel, a.Message, a.CreatedAt,
	)
	if err != nil {
		return Alert{}, fmt.Errorf("SaveAlert 失敗: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return Alert{}, fmt.Errorf("取得告警 ID 失敗: %w", err)
	}

	a.ID = id
	return a, nil
}

// UpdateAlertStatus 更新告警狀態，並記錄時間戳
func (s *SQLiteDB) UpdateAlertStatus(ctx context.Context, id int64, status AlertStatus) error {
	now := time.Now()
	var q string
	var args []any

	switch status {
	case AlertStatusSent:
		q = `UPDATE alerts SET status=?, sent_at=? WHERE id=?`
		args = []any{string(status), now, id}
	case AlertStatusAcked:
		q = `UPDATE alerts SET status=?, acked_at=? WHERE id=?`
		args = []any{string(status), now, id}
	default:
		q = `UPDATE alerts SET status=? WHERE id=?`
		args = []any{string(status), id}
	}

	if _, err := s.db.ExecContext(ctx, q, args...); err != nil {
		return fmt.Errorf("UpdateAlertStatus 失敗 id=%d: %w", id, err)
	}
	return nil
}

// QueryAlerts 查詢告警列表
func (s *SQLiteDB) QueryAlerts(ctx context.Context, status *AlertStatus, limit, offset int) ([]Alert, error) {
	q := `SELECT id, event_id, rule_id, status, channel, message, created_at, sent_at, acked_at FROM alerts`
	var args []any

	if status != nil {
		q += ` WHERE status=?`
		args = append(args, string(*status))
	}
	q += ` ORDER BY created_at DESC`
	if limit > 0 {
		q += fmt.Sprintf(` LIMIT %d OFFSET %d`, limit, offset)
	}

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("QueryAlerts 失敗: %w", err)
	}
	defer rows.Close()

	return scanAlerts(rows)
}

// ---- 規則 ----

// SaveRule 新增（ID=0）或更新規則
func (s *SQLiteDB) SaveRule(ctx context.Context, r Rule) (Rule, error) {
	if r.ID == 0 {
		// 新增
		const q = `INSERT INTO rules (name, description, enabled, cond_json, created_at) VALUES (?, ?, ?, ?, ?)`
		if r.CreatedAt.IsZero() {
			r.CreatedAt = time.Now()
		}
		if r.CondJSON == "" {
			r.CondJSON = "{}"
		}
		res, err := s.db.ExecContext(ctx, q, r.Name, r.Description, r.Enabled, r.CondJSON, r.CreatedAt)
		if err != nil {
			return Rule{}, fmt.Errorf("SaveRule insert 失敗: %w", err)
		}
		id, _ := res.LastInsertId()
		r.ID = id
		return r, nil
	}

	// 更新
	const q = `UPDATE rules SET name=?, description=?, enabled=?, cond_json=? WHERE id=?`
	if _, err := s.db.ExecContext(ctx, q, r.Name, r.Description, r.Enabled, r.CondJSON, r.ID); err != nil {
		return Rule{}, fmt.Errorf("SaveRule update 失敗 id=%d: %w", r.ID, err)
	}
	return r, nil
}

// DeleteRule 刪除規則
func (s *SQLiteDB) DeleteRule(ctx context.Context, id int64) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM rules WHERE id=?`, id); err != nil {
		return fmt.Errorf("DeleteRule 失敗 id=%d: %w", id, err)
	}
	return nil
}

// ListRules 列出規則
func (s *SQLiteDB) ListRules(ctx context.Context, enabledOnly bool) ([]Rule, error) {
	q := `SELECT id, name, description, enabled, cond_json, created_at FROM rules`
	if enabledOnly {
		q += ` WHERE enabled=1`
	}
	q += ` ORDER BY id ASC`

	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("ListRules 失敗: %w", err)
	}
	defer rows.Close()

	var rules []Rule
	for rows.Next() {
		var r Rule
		var enabled int
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &enabled, &r.CondJSON, &r.CreatedAt); err != nil {
			return nil, err
		}
		r.Enabled = enabled == 1
		rules = append(rules, r)
	}
	return rules, rows.Err()
}

// ---- 來源 ----

// SaveSource 新增（ID=0）或更新事件來源
func (s *SQLiteDB) SaveSource(ctx context.Context, src Source) (Source, error) {
	if src.ID == 0 {
		const q = `INSERT INTO sources (name, type, address, description, created_at) VALUES (?, ?, ?, ?, ?)`
		if src.CreatedAt.IsZero() {
			src.CreatedAt = time.Now()
		}
		res, err := s.db.ExecContext(ctx, q, src.Name, src.Type, src.Address, src.Description, src.CreatedAt)
		if err != nil {
			return Source{}, fmt.Errorf("SaveSource insert 失敗: %w", err)
		}
		id, _ := res.LastInsertId()
		src.ID = id
		return src, nil
	}

	const q = `UPDATE sources SET name=?, type=?, address=?, description=? WHERE id=?`
	if _, err := s.db.ExecContext(ctx, q, src.Name, src.Type, src.Address, src.Description, src.ID); err != nil {
		return Source{}, fmt.Errorf("SaveSource update 失敗 id=%d: %w", src.ID, err)
	}
	return src, nil
}

// ListSources 列出所有事件來源
func (s *SQLiteDB) ListSources(ctx context.Context) ([]Source, error) {
	const q = `SELECT id, name, type, address, description, created_at FROM sources ORDER BY id ASC`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("ListSources 失敗: %w", err)
	}
	defer rows.Close()

	var sources []Source
	for rows.Next() {
		var src Source
		if err := rows.Scan(&src.ID, &src.Name, &src.Type, &src.Address, &src.Description, &src.CreatedAt); err != nil {
			return nil, err
		}
		sources = append(sources, src)
	}
	return sources, rows.Err()
}

// DeleteSource 刪除事件來源
func (s *SQLiteDB) DeleteSource(ctx context.Context, id int64) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM sources WHERE id=?`, id); err != nil {
		return fmt.Errorf("DeleteSource 失敗 id=%d: %w", id, err)
	}
	return nil
}

// ListNotificationChannels 列出所有通知頻道
func (s *SQLiteDB) ListNotificationChannels(ctx context.Context) ([]NotificationChannel, error) {
	const q = `SELECT id, type, name, config_json, enabled, created_at FROM notification_channels ORDER BY id ASC`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("ListNotificationChannels 失敗: %w", err)
	}
	defer rows.Close()

	var channels []NotificationChannel
	for rows.Next() {
		var ch NotificationChannel
		var enabled int
		if err := rows.Scan(&ch.ID, &ch.Type, &ch.Name, &ch.ConfigJSON, &enabled, &ch.CreatedAt); err != nil {
			return nil, err
		}
		ch.Enabled = enabled == 1
		channels = append(channels, ch)
	}
	return channels, rows.Err()
}

// ---- 內部輔助函式 ----

// buildEventQuery 根據 EventFilter 動態組合 SQL
func buildEventQuery(filter EventFilter, count bool) (string, []any) {
	var where []string
	var args []any

	if filter.SourceID != nil {
		where = append(where, "source_id=?")
		args = append(args, *filter.SourceID)
	}
	if filter.Severity != nil {
		where = append(where, "severity=?")
		args = append(args, string(*filter.Severity))
	}
	if filter.From != nil {
		where = append(where, "parsed_at>=?")
		args = append(args, *filter.From)
	}
	if filter.To != nil {
		where = append(where, "parsed_at<=?")
		args = append(args, *filter.To)
	}

	var q string
	if count {
		q = "SELECT COUNT(*) FROM events"
	} else {
		q = "SELECT id, source_id, raw_log, parsed_at, severity, category, message, extra_json FROM events"
	}

	for i, w := range where {
		if i == 0 {
			q += " WHERE " + w
		} else {
			q += " AND " + w
		}
	}

	if !count {
		q += " ORDER BY parsed_at DESC"
		if filter.Limit > 0 {
			q += fmt.Sprintf(" LIMIT %d OFFSET %d", filter.Limit, filter.Offset)
		}
	}

	return q, args
}

// scanEvents 從 sql.Rows 掃描事件列表
func scanEvents(rows *sql.Rows) ([]Event, error) {
	var events []Event
	for rows.Next() {
		var e Event
		var severity string
		if err := rows.Scan(&e.ID, &e.SourceID, &e.RawLog, &e.ParsedAt,
			&severity, &e.Category, &e.Message, &e.ExtraJSON); err != nil {
			return nil, err
		}
		e.Severity = Severity(severity)
		events = append(events, e)
	}
	return events, rows.Err()
}

// scanAlerts 從 sql.Rows 掃描告警列表
func scanAlerts(rows *sql.Rows) ([]Alert, error) {
	var alerts []Alert
	for rows.Next() {
		var a Alert
		var status string
		if err := rows.Scan(&a.ID, &a.EventID, &a.RuleID, &status,
			&a.Channel, &a.Message, &a.CreatedAt, &a.SentAt, &a.AckedAt); err != nil {
			return nil, err
		}
		a.Status = AlertStatus(status)
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}
