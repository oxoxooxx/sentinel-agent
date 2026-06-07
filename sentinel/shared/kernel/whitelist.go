// Package kernel — 共用核心型別：白名單
package kernel

import "strings"

// Whitelist 代表一個字串白名單集合（共用型別）
// 可被多個 bounded context 引用（rule、event 等）
type Whitelist struct {
	items []string
}

// NewWhitelist 建立白名單，去除空白項目
func NewWhitelist(items []string) Whitelist {
	filtered := make([]string, 0, len(items))
	for _, item := range items {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			filtered = append(filtered, trimmed)
		}
	}
	return Whitelist{items: filtered}
}

// Contains 大小寫不敏感地檢查 value 是否在白名單內
func (w Whitelist) Contains(value string) bool {
	lower := strings.ToLower(value)
	for _, item := range w.items {
		if strings.Contains(lower, strings.ToLower(item)) {
			return true
		}
	}
	return false
}

// IsEmpty 檢查白名單是否為空
func (w Whitelist) IsEmpty() bool {
	return len(w.items) == 0
}

// Items 回傳白名單項目的唯讀副本
func (w Whitelist) Items() []string {
	result := make([]string, len(w.items))
	copy(result, w.items)
	return result
}

// Len 回傳白名單項目數量
func (w Whitelist) Len() int {
	return len(w.items)
}
