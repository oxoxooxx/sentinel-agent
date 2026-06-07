// Package api — 事件相關 API 處理器
package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	eventinfra "github.com/oxoxooxx/sentinel/internal/event/infra"
)

// EventsHandler 處理事件相關的 HTTP 請求
type EventsHandler struct {
	db eventinfra.DB
}

// NewEventsHandler 建立事件 handler
func NewEventsHandler(db eventinfra.DB) *EventsHandler {
	return &EventsHandler{db: db}
}

// List 處理 GET /api/events
// 支援查詢參數：severity, source_id, limit, offset
func (h *EventsHandler) List(c *gin.Context) {
	filter := eventinfra.EventFilter{
		Limit:  100, // 預設最多 100 筆
		Offset: 0,
	}

	// 解析 severity 篩選
	if s := c.Query("severity"); s != "" {
		sev := eventinfra.Severity(s)
		filter.Severity = &sev
	}

	// 解析 source_id 篩選
	if sid := c.Query("source_id"); sid != "" {
		id, err := strconv.ParseInt(sid, 10, 64)
		if err != nil {
			fail(c, http.StatusBadRequest, "source_id 必須是整數")
			return
		}
		filter.SourceID = &id
	}

	// 解析分頁參數
	if l := c.Query("limit"); l != "" {
		limit, err := strconv.Atoi(l)
		if err != nil || limit <= 0 || limit > 1000 {
			fail(c, http.StatusBadRequest, "limit 必須介於 1~1000")
			return
		}
		filter.Limit = limit
	}

	if o := c.Query("offset"); o != "" {
		offset, err := strconv.Atoi(o)
		if err != nil || offset < 0 {
			fail(c, http.StatusBadRequest, "offset 必須是非負整數")
			return
		}
		filter.Offset = offset
	}

	ctx := c.Request.Context()

	total, err := h.db.CountEvents(ctx, filter)
	if err != nil {
		fail(c, http.StatusInternalServerError, "查詢事件數量失敗")
		return
	}

	events, err := h.db.QueryEvents(ctx, filter)
	if err != nil {
		fail(c, http.StatusInternalServerError, "查詢事件失敗")
		return
	}

	// 避免回傳 null，統一回傳空陣列
	if events == nil {
		events = []eventinfra.Event{}
	}

	okWithTotal(c, events, total)
}

// Get 處理 GET /api/events/:id
func (h *EventsHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, "id 必須是整數")
		return
	}

	ctx := c.Request.Context()
	filter := eventinfra.EventFilter{
		Limit:  1,
		Offset: 0,
	}
	// 透過 QueryEvents 查詢單筆（TODO: 可加 FindByID 方法優化）
	events, err := h.db.QueryEvents(ctx, filter)
	if err != nil {
		fail(c, http.StatusInternalServerError, "查詢事件失敗")
		return
	}

	for _, e := range events {
		if e.ID == id {
			ok(c, e)
			return
		}
	}

	fail(c, http.StatusNotFound, "找不到指定事件")
}
