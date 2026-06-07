// Package api 提供 HTTP REST API 服務
package api

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oxoxooxx/sentinel/internal/storage"
)

// Server HTTP API 伺服器
type Server struct {
	db     storage.DB
	secret string
	router *gin.Engine
}

// NewServer 建立並設定 HTTP API 伺服器
func NewServer(db storage.DB, secret string) *Server {
	// 生產環境使用 Release 模式，減少 gin 的 debug log
	gin.SetMode(gin.ReleaseMode)

	s := &Server{
		db:     db,
		secret: secret,
		router: gin.New(),
	}

	s.setupMiddleware()
	s.setupRoutes()
	return s
}

// setupMiddleware 設定全域 middleware
func (s *Server) setupMiddleware() {
	// 結構化 log
	s.router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/health"},
	}))

	// panic recovery
	s.router.Use(gin.Recovery())

	// CORS（允許來自任何來源，適合內網部署；生產環境可限縮）
	s.router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})
}

// setupRoutes 設定 API 路由
func (s *Server) setupRoutes() {
	// 健康檢查（不需驗證）
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API v1 群組，需要 Bearer token 驗證
	v1 := s.router.Group("/api")
	if s.secret != "" {
		v1.Use(s.authMiddleware())
	}

	// 事件
	eventsHandler := NewEventsHandler(s.db)
	v1.GET("/events", eventsHandler.List)
	v1.GET("/events/:id", eventsHandler.Get)

	// 告警
	alertsHandler := NewAlertsHandler(s.db)
	v1.GET("/alerts", alertsHandler.List)
	v1.PUT("/alerts/:id/ack", alertsHandler.Ack)
	v1.PUT("/alerts/:id/resolve", alertsHandler.Resolve)

	// 規則
	rulesHandler := NewRulesHandler(s.db)
	v1.GET("/rules", rulesHandler.List)
	v1.POST("/rules", rulesHandler.Create)
	v1.PUT("/rules/:id", rulesHandler.Update)
	v1.DELETE("/rules/:id", rulesHandler.Delete)

	// 來源
	sourcesHandler := NewSourcesHandler(s.db)
	v1.GET("/sources", sourcesHandler.List)
	v1.POST("/sources", sourcesHandler.Create)
	v1.DELETE("/sources/:id", sourcesHandler.Delete)
}

// authMiddleware Bearer token 驗證 middleware
func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		expected := "Bearer " + s.secret
		if token != expected {
			slog.Warn("API 驗證失敗", "remote", c.ClientIP(), "path", c.Request.URL.Path)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "unauthorized",
			})
			return
		}
		c.Next()
	}
}

// Start 啟動 HTTP 伺服器，blocking 直到錯誤
func (s *Server) Start(port int) error {
	addr := fmt.Sprintf(":%d", port)
	slog.Info("API 伺服器啟動", "port", port)
	return s.router.Run(addr)
}

// APIResponse 統一 API 回應格式
type APIResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Total   int64  `json:"total,omitempty"`
}

// ok 回傳成功回應
func ok(c *gin.Context, data any) {
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: data})
}

// okWithTotal 回傳帶分頁資訊的成功回應
func okWithTotal(c *gin.Context, data any, total int64) {
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: data, Total: total})
}

// fail 回傳錯誤回應
func fail(c *gin.Context, statusCode int, msg string) {
	c.JSON(statusCode, APIResponse{Success: false, Error: msg})
}
