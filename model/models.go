package model

import "time"

type NotificationChannel struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	TelegramChatID string    `json:"telegram_chat_id"`
	CreatedAt      time.Time `json:"created_at"`
}

type Monitor struct {
	ID              string    `json:"id"`
	MonitorName     string    `json:"monitor_name"`
	CheckType       string    `json:"check_type"`
	Message         string    `json:"message"`
	Metadata        string    `json:"metadata"`
	Timeout         int       `json:"timeout"`
	ReAlertInterval int       `json:"re_alert_interval"`
	Status          string    `json:"status"`
	IsActive        bool      `json:"is_active"`
	ChannelID       *string   `json:"channel_id"`
	ServerIP        string    `json:"server_ip"`
	ServerName      string    `json:"server_name"`
	LastSeenAt      time.Time `json:"last_seen_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type AlertState struct {
	ID            string     `json:"id"`
	MonitorID     string     `json:"monitor_id"`
	Status        string     `json:"status"`
	LastAlertedAt time.Time  `json:"last_alerted_at"`
	FiredAt       time.Time  `json:"fired_at"`
	ResolvedAt    *time.Time `json:"resolved_at,omitempty"`
}

type NotificationLog struct {
	ID        string    `json:"id"`
	MonitorID string    `json:"monitor_id"`
	ChannelID *string   `json:"channel_id"`
	AlertType string    `json:"alert_type"` // "alert", "re_alert", "recovered"
	Message   string    `json:"message"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
