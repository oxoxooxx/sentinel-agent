// Package domain — source 領域層：Source aggregate
package domain

import (
	"fmt"
	"time"
)

// SourceType 事件來源類型
type SourceType string

const (
	SourceTypeFortiGate SourceType = "fortigate" // FortiGate 防火牆
	SourceTypeNAS       SourceType = "nas"       // NAS 儲存設備
	SourceTypeWebhook   SourceType = "webhook"   // HTTP webhook
)

// Source 代表一個事件來源（aggregate root）
type Source struct {
	ID          int64
	Name        string
	Type        SourceType
	Address     string // IP 位址或 URL
	Description string
	CreatedAt   time.Time
}

// NewSource 建立新的事件來源
func NewSource(name string, srcType SourceType, address, description string) (Source, error) {
	if name == "" {
		return Source{}, fmt.Errorf("來源名稱不可為空")
	}
	if err := validateSourceType(srcType); err != nil {
		return Source{}, err
	}

	return Source{
		Name:        name,
		Type:        srcType,
		Address:     address,
		Description: description,
		CreatedAt:   time.Now(),
	}, nil
}

// validateSourceType 驗證來源類型是否合法
func validateSourceType(t SourceType) error {
	switch t {
	case SourceTypeFortiGate, SourceTypeNAS, SourceTypeWebhook:
		return nil
	}
	return fmt.Errorf("不支援的來源類型: %s（支援: fortigate, nas, webhook）", t)
}
