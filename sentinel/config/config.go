// Package config 負責載入並解析 YAML 設定檔
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 是整個 Sentinel 服務的根設定
type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Storage StorageConfig `yaml:"storage"`
	Syslog  SyslogConfig  `yaml:"syslog"`
	API     APIConfig     `yaml:"api"`
	Alert   AlertConfig   `yaml:"alert"`
}

// ServerConfig 服務基本資訊
type ServerConfig struct {
	Name string `yaml:"name"`
}

// StorageConfig 儲存後端設定（支援 sqlite 或 postgres）
type StorageConfig struct {
	Driver   string        `yaml:"driver"` // "sqlite" 或 "postgres"
	SQLite   SQLiteConfig  `yaml:"sqlite"`
	Postgres PostgresConfig `yaml:"postgres"`
}

// SQLiteConfig SQLite 專用設定
type SQLiteConfig struct {
	Path          string `yaml:"path"`           // 資料庫檔案路徑
	RetentionDays int    `yaml:"retention_days"` // 事件保留天數
}

// PostgresConfig PostgreSQL 專用設定
type PostgresConfig struct {
	DSN string `yaml:"dsn"` // postgres://user:pass@host:5432/db
}

// SyslogConfig UDP syslog 接收器設定
type SyslogConfig struct {
	Port     int    `yaml:"port"`     // 預設 514
	Protocol string `yaml:"protocol"` // "udp" 或 "tcp"
}

// APIConfig HTTP API 伺服器設定
type APIConfig struct {
	Port   int    `yaml:"port"`   // 預設 8080
	Secret string `yaml:"secret"` // Bearer token，用於保護 API
}

// AlertConfig 告警派送設定
type AlertConfig struct {
	Channels []ChannelConfig `yaml:"channels"`
}

// ChannelConfig 單一通知頻道設定
type ChannelConfig struct {
	Type string `yaml:"type"` // "line", "telegram", "teams"

	// LINE Notify
	Token string `yaml:"token,omitempty"`

	// Telegram Bot
	BotToken string `yaml:"bot_token,omitempty"`
	ChatID   string `yaml:"chat_id,omitempty"`

	// Microsoft Teams
	WebhookURL string `yaml:"webhook_url,omitempty"`
}

// Load 從指定路徑載入 YAML 設定檔，並套用預設值
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀取設定檔失敗 %s: %w", path, err)
	}

	cfg := defaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("解析設定檔失敗: %w", err)
	}

	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("設定驗證失敗: %w", err)
	}

	return cfg, nil
}

// defaultConfig 回傳含預設值的設定
func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Name: "Sentinel",
		},
		Storage: StorageConfig{
			Driver: "sqlite",
			SQLite: SQLiteConfig{
				Path:          "./data/sentinel.db",
				RetentionDays: 90,
			},
		},
		Syslog: SyslogConfig{
			Port:     514,
			Protocol: "udp",
		},
		API: APIConfig{
			Port: 8080,
		},
	}
}

// validate 檢查必要欄位
func validate(cfg *Config) error {
	switch cfg.Storage.Driver {
	case "sqlite", "postgres":
		// 合法
	default:
		return fmt.Errorf("不支援的 storage driver: %s（請用 sqlite 或 postgres）", cfg.Storage.Driver)
	}

	if cfg.Storage.Driver == "sqlite" && cfg.Storage.SQLite.Path == "" {
		return fmt.Errorf("sqlite.path 不可為空")
	}

	if cfg.Storage.Driver == "postgres" && cfg.Storage.Postgres.DSN == "" {
		return fmt.Errorf("postgres.dsn 不可為空")
	}

	if cfg.Syslog.Port <= 0 || cfg.Syslog.Port > 65535 {
		return fmt.Errorf("syslog.port 無效: %d", cfg.Syslog.Port)
	}

	if cfg.API.Port <= 0 || cfg.API.Port > 65535 {
		return fmt.Errorf("api.port 無效: %d", cfg.API.Port)
	}

	return nil
}
