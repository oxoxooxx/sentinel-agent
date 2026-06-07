// Package infra — alert 基礎設施層：Telegram Bot 通知頻道
package infra

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TelegramNotifier 透過 Telegram Bot API 發送告警訊息
type TelegramNotifier struct {
	botToken string
	chatID   string
	client   *http.Client
}

// NewTelegramNotifier 建立 Telegram 通知器
func NewTelegramNotifier(botToken, chatID string) (*TelegramNotifier, error) {
	if botToken == "" {
		return nil, fmt.Errorf("Telegram bot_token 不可為空")
	}
	if chatID == "" {
		return nil, fmt.Errorf("Telegram chat_id 不可為空")
	}
	return &TelegramNotifier{
		botToken: botToken,
		chatID:   chatID,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// Name 回傳頻道名稱
func (n *TelegramNotifier) Name() string {
	return "telegram"
}

// Send 透過 Telegram sendMessage API 發送訊息
func (n *TelegramNotifier) Send(ctx context.Context, msg string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.botToken)

	payload := map[string]string{
		"chat_id": n.chatID,
		"text":    msg,
		// 使用 HTML parse mode 支援基本排版
		"parse_mode": "HTML",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化 Telegram 請求失敗: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("建立 Telegram 請求失敗: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("Telegram 請求失敗: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Telegram API 回傳錯誤 HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
