package fileMonitor

import (
	"encoding/json"
	"fmt"
	"time"
)

// 告警类型
type AlertType int

const (
	AlertSuspicious AlertType = iota // ⚠️ 可疑行为
	AlertHighRisk                    // ❌ 高危行为（黑名单）
)

type Alert struct {
	Type      AlertType `json:"type"`
	Path      string    `json:"path"`
	Timestamp time.Time `json:"timestamp"`
	Details   string    `json:"details"`
}

// 告警生成器接口
type AlertGenerator interface {
	GenerateAlert(alertType AlertType, path string) string
}

type DefaultAlertGenerator struct{}

func NewAlertGenerator() AlertGenerator {
	return &DefaultAlertGenerator{}
}

func (ag *DefaultAlertGenerator) GenerateAlert(alertType AlertType, path string) string {
	alert := Alert{
		Type:      alertType,
		Path:      path,
		Timestamp: time.Now(),
	}

	switch alertType {
	case AlertSuspicious:
		alert.Details = fmt.Sprintf("Suspicious activity detected on %s", path)
	case AlertHighRisk:
		alert.Details = fmt.Sprintf("High-risk activity detected (blacklisted file) on %s", path)
	default:
		alert.Details = "Unknown alert"
	}

	alertJSON, err := json.Marshal(alert)
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to generate alert: %v"}`, err)
	}

	return string(alertJSON)
}
