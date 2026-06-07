// Package api — 告警相關 API 處理器
package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	eventinfra "github.com/oxoxooxx/sentinel/internal/event/infra"
)

// AlertsHandler 處理告警相關的 HTTP 請求
type AlertsHandler struct {
	db eventinfra.DB
}

// NewAlertsHandler 建立告警 handler
func NewAlertsHandler(db eventinfra.DB) *AlertsHandler {
	return &AlertsHandler{db: db}
}

// List 處理 GET /api/alerts
// 支援查詢參數：status, limit, offset
func (h *AlertsHandler) List(c *gin.Context) {
	var statusFilter *eventinfra.AlertStatus
	if s := c.Query("status"); s != "" {
		st := eventinfra.AlertStatus(s)
		statusFilter = &st
	}

	limit := 100
	offset := 0

	if l := c.Query("limit"); l != "" {
		v, err := strconv.Atoi(l)
		if err != nil || v <= 0 || v > 1000 {
			fail(c, http.StatusBadRequest, "limit 必須介於 1~1000")
			return
		}
		limit = v
	}

	if o := c.Query("offset"); o != "" {
		v, err := strconv.Atoi(o)
		if err != nil || v < 0 {
			fail(c, http.StatusBadRequest, "offset 必須是非負整數")
			return
		}
		offset = v
	}

	ctx := c.Request.Context()
	alerts, err := h.db.QueryAlerts(ctx, statusFilter, limit, offset)
	if err != nil {
		fail(c, http.StatusInternalServerError, "查詢告警失敗")
		return
	}

	if alerts == nil {
		alerts = []eventinfra.Alert{}
	}

	ok(c, alerts)
}

// Ack 處理 PUT /api/alerts/:id/ack（確認告警）
func (h *AlertsHandler) Ack(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, "id 必須是整數")
		return
	}

	ctx := c.Request.Context()
	if err := h.db.UpdateAlertStatus(ctx, id, eventinfra.AlertStatusAcked); err != nil {
		fail(c, http.StatusInternalServerError, "更新告警狀態失敗")
		return
	}

	ok(c, gin.H{"id": id, "status": eventinfra.AlertStatusAcked})
}

// Resolve 處理 PUT /api/alerts/:id/resolve（解決告警）
func (h *AlertsHandler) Resolve(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, "id 必須是整數")
		return
	}

	ctx := c.Request.Context()
	if err := h.db.UpdateAlertStatus(ctx, id, eventinfra.AlertStatusResolved); err != nil {
		fail(c, http.StatusInternalServerError, "更新告警狀態失敗")
		return
	}

	ok(c, gin.H{"id": id, "status": eventinfra.AlertStatusResolved})
}
