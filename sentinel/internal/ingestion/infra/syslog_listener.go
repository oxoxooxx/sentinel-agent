// Package infra — ingestion 基礎設施層：UDP syslog 監聽器
package infra

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	eventinfra "github.com/oxoxooxx/sentinel/internal/event/infra"
)

// SyslogMessage 代表一筆解析後的 syslog 訊息
type SyslogMessage struct {
	ReceivedAt time.Time
	RemoteAddr string
	RawData    string
}

// SyslogHandler 是接收到 syslog 訊息後的回呼介面
type SyslogHandler interface {
	HandleSyslog(ctx context.Context, msg SyslogMessage) error
}

// SyslogServer 監聽 UDP syslog（主要對應 FortiGate）
type SyslogServer struct {
	port    int
	handler SyslogHandler
	conn    *net.UDPConn
}

// NewSyslogServer 建立 UDP syslog 監聽器
func NewSyslogServer(port int, handler SyslogHandler) *SyslogServer {
	return &SyslogServer{
		port:    port,
		handler: handler,
	}
}

// Start 開始監聽，直到 ctx 取消為止
// 此函式為 blocking，請在 goroutine 中呼叫
func (s *SyslogServer) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", s.port)
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		return fmt.Errorf("syslog UDP 監聽失敗 %s: %w", addr, err)
	}
	s.conn = conn.(*net.UDPConn)

	slog.Info("syslog 監聽啟動", "port", s.port, "protocol", "udp")

	// 透過 goroutine 偵測 ctx 取消，關閉 conn 以中斷 ReadFrom
	go func() {
		<-ctx.Done()
		s.conn.Close()
	}()

	return s.listenLoop(ctx)
}

// listenLoop 持續接收 UDP 封包並派發給 handler
func (s *SyslogServer) listenLoop(ctx context.Context) error {
	// syslog 最大單筆訊息 64KB，保留緩衝
	buf := make([]byte, 65535)

	for {
		n, remoteAddr, err := s.conn.ReadFrom(buf)
		if err != nil {
			// ctx 取消時 conn 被關閉，屬於正常退出
			select {
			case <-ctx.Done():
				slog.Info("syslog 監聽停止")
				return nil
			default:
				return fmt.Errorf("syslog 讀取失敗: %w", err)
			}
		}

		// 複製一份 buffer 以避免資料競爭
		raw := string(buf[:n])
		msg := SyslogMessage{
			ReceivedAt: time.Now(),
			RemoteAddr: remoteAddr.String(),
			RawData:    raw,
		}

		// 非同步處理，不阻塞接收迴圈
		go func(m SyslogMessage) {
			if err := s.handler.HandleSyslog(ctx, m); err != nil {
				slog.Error("syslog 處理失敗", "remote", m.RemoteAddr, "err", err)
			}
		}(msg)
	}
}

// DefaultSyslogHandler 預設的 syslog 處理器，將訊息存入 DB
type DefaultSyslogHandler struct {
	db       eventinfra.DB
	sourceID int64 // 預設來源 ID（FortiGate）
}

// NewDefaultSyslogHandler 建立預設 syslog 處理器
func NewDefaultSyslogHandler(db eventinfra.DB, sourceID int64) *DefaultSyslogHandler {
	return &DefaultSyslogHandler{db: db, sourceID: sourceID}
}

// HandleSyslog 解析並儲存 syslog 訊息
func (h *DefaultSyslogHandler) HandleSyslog(ctx context.Context, msg SyslogMessage) error {
	// TODO: 實作 FortiGate syslog 格式解析（RFC 5424 / CEF）
	// 目前先直接儲存原始資料，後續版本加入 parser
	severity, message := parseSeverity(msg.RawData)

	event := eventinfra.Event{
		SourceID:  h.sourceID,
		RawLog:    msg.RawData,
		ParsedAt:  msg.ReceivedAt,
		Severity:  severity,
		Category:  "syslog",
		Message:   message,
		ExtraJSON: fmt.Sprintf(`{"remote_addr": "%s"}`, msg.RemoteAddr),
	}

	if _, err := h.db.SaveEvent(ctx, event); err != nil {
		return fmt.Errorf("儲存 syslog 事件失敗: %w", err)
	}
	return nil
}

// parseSeverity 從 syslog 原始訊息中推斷嚴重等級
// 這是簡化實作，後續可擴充為完整的 RFC 5424 parser
func parseSeverity(raw string) (eventinfra.Severity, string) {
	// FortiGate syslog 常見關鍵字判斷
	keywords := map[string]eventinfra.Severity{
		"level=emergency": eventinfra.SeverityCritical,
		"level=alert":     eventinfra.SeverityCritical,
		"level=critical":  eventinfra.SeverityCritical,
		"level=error":     eventinfra.SeverityWarning,
		"level=warning":   eventinfra.SeverityWarning,
	}

	for kw, sev := range keywords {
		if containsIgnoreCase(raw, kw) {
			return sev, raw
		}
	}

	return eventinfra.SeverityInfo, raw
}

// containsIgnoreCase 大小寫不敏感的字串搜尋
func containsIgnoreCase(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			c1, c2 := s[i+j], substr[j]
			// 轉小寫比較（僅處理 ASCII）
			if c1 >= 'A' && c1 <= 'Z' {
				c1 += 32
			}
			if c2 >= 'A' && c2 <= 'Z' {
				c2 += 32
			}
			if c1 != c2 {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
