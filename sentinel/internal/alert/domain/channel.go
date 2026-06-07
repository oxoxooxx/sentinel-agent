// Package domain — alert 領域層：NotificationChannel value object
package domain

import "fmt"

// ChannelType 通知頻道類型
type ChannelType string

const (
	ChannelTypeLINE     ChannelType = "line"     // LINE Notify
	ChannelTypeTelegram ChannelType = "telegram" // Telegram Bot
	ChannelTypeTeams    ChannelType = "teams"    // Microsoft Teams
)

// NotificationChannel 代表一個通知頻道設定（value object）
type NotificationChannel struct {
	ID         int64
	Type       ChannelType
	Name       string
	ConfigJSON string // 頻道設定（JSON 字串，含 token 等機密）
	Enabled    bool
}

// IsValid 檢查頻道設定是否合法
func (c NotificationChannel) IsValid() bool {
	switch c.Type {
	case ChannelTypeLINE, ChannelTypeTelegram, ChannelTypeTeams:
		return c.Name != ""
	}
	return false
}

// Validate 驗證頻道類型，不合法時回傳錯誤
func ValidateChannelType(t string) error {
	switch ChannelType(t) {
	case ChannelTypeLINE, ChannelTypeTelegram, ChannelTypeTeams:
		return nil
	}
	return fmt.Errorf("不支援的通知頻道類型: %s（支援: line, telegram, teams）", t)
}
