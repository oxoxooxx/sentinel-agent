// Package alert — Microsoft Teams Webhook 通知頻道
package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TeamsNotifier 透過 Microsoft Teams Incoming Webhook 發送告警訊息
type TeamsNotifier struct {
	webhookURL string
	client     *http.Client
}

// NewTeamsNotifier 建立 Teams 通知器
func NewTeamsNotifier(webhookURL string) (*TeamsNotifier, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("Teams webhook_url 不可為空")
	}
	return &TeamsNotifier{
		webhookURL: webhookURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// Name 回傳頻道名稱
func (n *TeamsNotifier) Name() string {
	return "teams"
}

// teamsPayload 是 Teams Incoming Webhook 的 Adaptive Card 格式
// 使用 MessageCard 格式（較廣泛支援）
type teamsPayload struct {
	Type       string `json:"@type"`
	Context    string `json:"@context"`
	ThemeColor string `json:"themeColor"`
	Summary    string `json:"summary"`
	Sections   []teamsSection `json:"sections"`
}

type teamsSection struct {
	ActivityTitle string `json:"activityTitle"`
	ActivityText  string `json:"activityText"`
}

// Send 透過 Teams Incoming Webhook 發送訊息
func (n *TeamsNotifier) Send(ctx context.Context, msg string) error {
	payload := teamsPayload{
		Type:       "MessageCard",
		Context:    "http://schema.org/extensions",
		ThemeColor: "FF0000", // 紅色代表告警
		Summary:    "Sentinel 安全告警",
		Sections: []teamsSection{
			{
				ActivityTitle: "Sentinel 安全告警",
				ActivityText:  msg,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化 Teams 請求失敗: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("建立 Teams 請求失敗: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("Teams 請求失敗: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Teams Webhook 回傳錯誤 HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
