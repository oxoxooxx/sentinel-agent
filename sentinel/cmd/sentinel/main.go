// Sentinel — 輕量資安監控平台入口
// 啟動流程：載入設定 → 初始化 DB → 啟動 syslog 接收 → 啟動 webhook → 啟動 API 伺服器
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/oxoxooxx/sentinel/config"
	"github.com/oxoxooxx/sentinel/internal/alert"
	"github.com/oxoxooxx/sentinel/internal/api"
	"github.com/oxoxooxx/sentinel/internal/ingestion"
	"github.com/oxoxooxx/sentinel/internal/storage"
)

func main() {
	// 結構化 log，輸出到 stdout
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// 讀取設定檔路徑（預設 config.yaml，可透過環境變數覆寫）
	cfgPath := getEnv("SENTINEL_CONFIG", "config.yaml")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		slog.Error("載入設定失敗", "path", cfgPath, "err", err)
		os.Exit(1)
	}

	slog.Info("Sentinel 啟動中", "name", cfg.Server.Name, "storage", cfg.Storage.Driver)

	// 初始化儲存後端
	db, err := initStorage(cfg.Storage)
	if err != nil {
		slog.Error("初始化儲存後端失敗", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	// 執行資料庫 schema migration
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := db.Migrate(ctx); err != nil {
		slog.Error("資料庫 migration 失敗", "err", err)
		os.Exit(1)
	}
	slog.Info("資料庫 schema 就緒")

	// 初始化告警派送器
	dispatcher, err := alert.NewDispatcher(db, cfg.Alert)
	if err != nil {
		slog.Error("初始化告警派送器失敗", "err", err)
		os.Exit(1)
	}
	_ = dispatcher // 後續版本接入規則引擎

	// 啟動 UDP syslog 監聽（背景 goroutine）
	go func() {
		// sourceID=1 預設對應第一個 FortiGate 來源
		// 實際部署時應從 DB 查詢對應的 source ID
		handler := ingestion.NewDefaultSyslogHandler(db, 1)
		server := ingestion.NewSyslogServer(cfg.Syslog.Port, handler)
		if err := server.Start(ctx); err != nil {
			slog.Error("syslog 伺服器錯誤", "err", err)
		}
	}()

	// 啟動 HTTP API 伺服器（背景 goroutine）
	go func() {
		apiServer := api.NewServer(db, cfg.API.Secret)
		if err := apiServer.Start(cfg.API.Port); err != nil {
			slog.Error("API 伺服器錯誤", "err", err)
		}
	}()

	slog.Info("所有服務啟動完成",
		"syslog_port", cfg.Syslog.Port,
		"api_port", cfg.API.Port,
	)

	// 等待 SIGINT 或 SIGTERM 信號，優雅關閉
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	slog.Info("收到關閉信號，正在停止服務...", "signal", sig.String())
	cancel() // 通知所有 goroutine 停止
	slog.Info("Sentinel 已停止")
}

// initStorage 根據設定初始化儲存後端
func initStorage(cfg config.StorageConfig) (storage.DB, error) {
	switch cfg.Driver {
	case "sqlite":
		// 確保資料目錄存在
		if err := os.MkdirAll(getDir(cfg.SQLite.Path), 0750); err != nil {
			slog.Warn("無法建立資料目錄", "path", cfg.SQLite.Path, "err", err)
		}
		return storage.NewSQLite(cfg.SQLite.Path)
	case "postgres":
		// TODO: 實作 PostgreSQL 後端
		slog.Error("PostgreSQL 後端尚未實作，請使用 sqlite")
		os.Exit(1)
		return nil, nil
	default:
		slog.Error("不支援的 storage driver", "driver", cfg.Driver)
		os.Exit(1)
		return nil, nil
	}
}

// getEnv 取得環境變數，若不存在則回傳預設值
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getDir 取得檔案路徑的目錄部分
func getDir(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[:i]
		}
	}
	return "."
}
