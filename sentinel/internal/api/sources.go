// Package api — 事件來源 API 處理器
package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/oxoxooxx/sentinel/internal/storage"
)

// SourcesHandler 處理事件來源 CRUD 請求
type SourcesHandler struct {
	db storage.DB
}

// NewSourcesHandler 建立來源 handler
func NewSourcesHandler(db storage.DB) *SourcesHandler {
	return &SourcesHandler{db: db}
}

// createSourceRequest 新增事件來源的請求 body
type createSourceRequest struct {
	Name        string `json:"name"    binding:"required"`
	Type        string `json:"type"    binding:"required"`  // fortigate | nas | webhook
	Address     string `json:"address"`
	Description string `json:"description"`
}

// List 處理 GET /api/sources
func (h *SourcesHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	sources, err := h.db.ListSources(ctx)
	if err != nil {
		fail(c, http.StatusInternalServerError, "查詢來源失敗")
		return
	}

	if sources == nil {
		sources = []storage.Source{}
	}

	ok(c, sources)
}

// Create 處理 POST /api/sources
func (h *SourcesHandler) Create(c *gin.Context) {
	var req createSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "請求格式錯誤: "+err.Error())
		return
	}

	// 驗證 type 欄位
	validTypes := map[string]bool{"fortigate": true, "nas": true, "webhook": true}
	if !validTypes[req.Type] {
		fail(c, http.StatusBadRequest, "type 必須是 fortigate、nas 或 webhook")
		return
	}

	source := storage.Source{
		Name:        req.Name,
		Type:        req.Type,
		Address:     req.Address,
		Description: req.Description,
	}

	ctx := c.Request.Context()
	saved, err := h.db.SaveSource(ctx, source)
	if err != nil {
		fail(c, http.StatusInternalServerError, "建立來源失敗")
		return
	}

	c.JSON(http.StatusCreated, APIResponse{Success: true, Data: saved})
}

// Delete 處理 DELETE /api/sources/:id
func (h *SourcesHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, "id 必須是整數")
		return
	}

	ctx := c.Request.Context()
	if err := h.db.DeleteSource(ctx, id); err != nil {
		fail(c, http.StatusInternalServerError, "刪除來源失敗")
		return
	}

	ok(c, gin.H{"deleted": id})
}
