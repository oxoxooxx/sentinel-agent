// Package alert — LINE Notify 通知頻道
package alert

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const lineNotifyURL = "https://notify-api.line.me/api/notify"

// LINENotifier 透過 LINE Notify API 發送告警訊息
type LINENotifier struct {
	token  string
	client *http.Client
}

// NewLINENotifier 建立 LINE Notify 通知器
// token 為 LINE Notify 的 Personal Access Token
func NewLINENotifier(token string) (*LINENotifier, error) {
	if token == "" {
		return nil, fmt.Errorf("LINE Notify token 不可為空")
	}
	return &LINENotifier{
		token: token,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// Name 回傳頻道名稱
func (n *LINENotifier) Name() string {
	return "line"
}

// Send 透過 LINE Notify API 發送訊息
func (n *LINENotifier) Send(ctx context.Context, msg string) error {
	// LINE Notify 使用 form-encoded POST
	form := url.Values{}
	form.Set("message", msg)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, lineNotifyURL,
		strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("建立 LINE Notify 請求失敗: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+n.token)

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("LINE Notify 請求失敗: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("LINE Notify 回傳錯誤 HTTP %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
