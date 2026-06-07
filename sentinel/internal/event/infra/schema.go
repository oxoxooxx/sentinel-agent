// Package infra — event 基礎設施層：資料庫 schema 定義
package infra

// SQLiteSchema 包含建立所有資料表的 SQL 語句（SQLite 方言）
// 使用 IF NOT EXISTS，可安全重複執行
const SQLiteSchema = `
-- 事件來源（FortiGate、NAS、webhook 等）
CREATE TABLE IF NOT EXISTS sources (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT    NOT NULL,
    type        TEXT    NOT NULL,      -- fortigate | nas | webhook
    address     TEXT    NOT NULL DEFAULT '',
    description TEXT    NOT NULL DEFAULT '',
    created_at  DATETIME NOT NULL DEFAULT (datetime('now'))
);

-- 安全事件（原始 log 解析後的記錄）
CREATE TABLE IF NOT EXISTS events (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    source_id   INTEGER NOT NULL,
    raw_log     TEXT    NOT NULL,
    parsed_at   DATETIME NOT NULL DEFAULT (datetime('now')),
    severity    TEXT    NOT NULL DEFAULT 'info',  -- info | warning | critical
    category    TEXT    NOT NULL DEFAULT '',
    message     TEXT    NOT NULL DEFAULT '',
    extra_json  TEXT    NOT NULL DEFAULT '{}',
    FOREIGN KEY (source_id) REFERENCES sources(id) ON DELETE CASCADE
);

-- 加速常見查詢
CREATE INDEX IF NOT EXISTS idx_events_parsed_at  ON events(parsed_at DESC);
CREATE INDEX IF NOT EXISTS idx_events_severity   ON events(severity);
CREATE INDEX IF NOT EXISTS idx_events_source_id  ON events(source_id);

-- 告警規則
CREATE TABLE IF NOT EXISTS rules (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT    NOT NULL,
    description TEXT    NOT NULL DEFAULT '',
    enabled     INTEGER NOT NULL DEFAULT 1,  -- 1=啟用, 0=停用
    cond_json   TEXT    NOT NULL DEFAULT '{}',
    created_at  DATETIME NOT NULL DEFAULT (datetime('now'))
);

-- 告警記錄
CREATE TABLE IF NOT EXISTS alerts (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id    INTEGER NOT NULL,
    rule_id     INTEGER NOT NULL,
    status      TEXT    NOT NULL DEFAULT 'pending', -- pending | sent | acked | resolved
    channel     TEXT    NOT NULL DEFAULT '',
    message     TEXT    NOT NULL DEFAULT '',
    created_at  DATETIME NOT NULL DEFAULT (datetime('now')),
    sent_at     DATETIME,
    acked_at    DATETIME,
    FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
    FOREIGN KEY (rule_id)  REFERENCES rules(id)  ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_alerts_status     ON alerts(status);
CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts(created_at DESC);

-- 通知頻道設定（LINE / Telegram / Teams）
CREATE TABLE IF NOT EXISTS notification_channels (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    type        TEXT    NOT NULL,  -- line | telegram | teams
    name        TEXT    NOT NULL,
    config_json TEXT    NOT NULL DEFAULT '{}',  -- 加密儲存 token 等機密
    enabled     INTEGER NOT NULL DEFAULT 1,
    created_at  DATETIME NOT NULL DEFAULT (datetime('now'))
);
`

// PostgresSchema 包含建立所有資料表的 SQL 語句（PostgreSQL 方言）
const PostgresSchema = `
-- 事件來源
CREATE TABLE IF NOT EXISTS sources (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT      NOT NULL,
    type        TEXT      NOT NULL,
    address     TEXT      NOT NULL DEFAULT '',
    description TEXT      NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 安全事件
CREATE TABLE IF NOT EXISTS events (
    id          BIGSERIAL PRIMARY KEY,
    source_id   BIGINT    NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
    raw_log     TEXT      NOT NULL,
    parsed_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    severity    TEXT      NOT NULL DEFAULT 'info',
    category    TEXT      NOT NULL DEFAULT '',
    message     TEXT      NOT NULL DEFAULT '',
    extra_json  JSONB     NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_events_parsed_at  ON events(parsed_at DESC);
CREATE INDEX IF NOT EXISTS idx_events_severity   ON events(severity);
CREATE INDEX IF NOT EXISTS idx_events_source_id  ON events(source_id);

-- 告警規則
CREATE TABLE IF NOT EXISTS rules (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT      NOT NULL,
    description TEXT      NOT NULL DEFAULT '',
    enabled     BOOLEAN   NOT NULL DEFAULT TRUE,
    cond_json   JSONB     NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 告警記錄
CREATE TABLE IF NOT EXISTS alerts (
    id          BIGSERIAL PRIMARY KEY,
    event_id    BIGINT    NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    rule_id     BIGINT    NOT NULL REFERENCES rules(id)  ON DELETE CASCADE,
    status      TEXT      NOT NULL DEFAULT 'pending',
    channel     TEXT      NOT NULL DEFAULT '',
    message     TEXT      NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sent_at     TIMESTAMPTZ,
    acked_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_alerts_status     ON alerts(status);
CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts(created_at DESC);

-- 通知頻道設定
CREATE TABLE IF NOT EXISTS notification_channels (
    id          BIGSERIAL PRIMARY KEY,
    type        TEXT      NOT NULL,
    name        TEXT      NOT NULL,
    config_json JSONB     NOT NULL DEFAULT '{}',
    enabled     BOOLEAN   NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`
