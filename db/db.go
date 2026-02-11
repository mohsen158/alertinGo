package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mohsen/alertinGo/model"
)

var Pool *pgxpool.Pool

func Connect() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	var err error
	Pool, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}

	if err := Pool.Ping(context.Background()); err != nil {
		log.Fatalf("unable to ping database: %v", err)
	}

	log.Println("connected to database")
}

func RunMigrations() {
	sql, err := os.ReadFile("migrations/001_initial.sql")
	if err != nil {
		log.Fatalf("failed to read migration file: %v", err)
	}

	if _, err := Pool.Exec(context.Background(), string(sql)); err != nil {
		log.Printf("migration warning (may already be applied): %v", err)
	} else {
		log.Println("migrations applied")
	}
}

// --- Monitors ---

func UpsertMonitor(ctx context.Context, m *model.Monitor) (*model.Monitor, error) {
	query := `
		INSERT INTO monitors (monitor_name, check_type, message, metadata, timeout, re_alert_interval, server_ip, server_name, last_seen_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now(), now())
		ON CONFLICT (monitor_name, check_type)
		DO UPDATE SET
			message = EXCLUDED.message,
			metadata = EXCLUDED.metadata,
			timeout = EXCLUDED.timeout,
			re_alert_interval = EXCLUDED.re_alert_interval,
			server_ip = EXCLUDED.server_ip,
			server_name = EXCLUDED.server_name,
			last_seen_at = now(),
			updated_at = now(),
			status = 'up'
		RETURNING id, monitor_name, check_type, message, metadata, timeout, re_alert_interval, status, is_active, channel_id, server_ip, server_name, last_seen_at, created_at, updated_at`

	var mon model.Monitor
	err := Pool.QueryRow(ctx, query,
		m.MonitorName, m.CheckType, m.Message, m.Metadata,
		m.Timeout, m.ReAlertInterval, m.ServerIP, m.ServerName,
	).Scan(
		&mon.ID, &mon.MonitorName, &mon.CheckType, &mon.Message, &mon.Metadata,
		&mon.Timeout, &mon.ReAlertInterval, &mon.Status, &mon.IsActive, &mon.ChannelID,
		&mon.ServerIP, &mon.ServerName, &mon.LastSeenAt, &mon.CreatedAt, &mon.UpdatedAt,
	)
	return &mon, err
}

func GetAllMonitors(ctx context.Context) ([]model.Monitor, error) {
	query := `SELECT id, monitor_name, check_type, message, metadata, timeout, re_alert_interval, status, is_active, channel_id, server_ip, server_name, last_seen_at, created_at, updated_at FROM monitors ORDER BY created_at DESC`

	rows, err := Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var monitors []model.Monitor
	for rows.Next() {
		var m model.Monitor
		if err := rows.Scan(&m.ID, &m.MonitorName, &m.CheckType, &m.Message, &m.Metadata,
			&m.Timeout, &m.ReAlertInterval, &m.Status, &m.IsActive, &m.ChannelID,
			&m.ServerIP, &m.ServerName, &m.LastSeenAt, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		monitors = append(monitors, m)
	}
	return monitors, nil
}

func GetMonitorByID(ctx context.Context, id string) (*model.Monitor, error) {
	query := `SELECT id, monitor_name, check_type, message, metadata, timeout, re_alert_interval, status, is_active, channel_id, server_ip, server_name, last_seen_at, created_at, updated_at FROM monitors WHERE id = $1`

	var m model.Monitor
	err := Pool.QueryRow(ctx, query, id).Scan(
		&m.ID, &m.MonitorName, &m.CheckType, &m.Message, &m.Metadata,
		&m.Timeout, &m.ReAlertInterval, &m.Status, &m.IsActive, &m.ChannelID,
		&m.ServerIP, &m.ServerName, &m.LastSeenAt, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func UpdateMonitor(ctx context.Context, id string, isActive bool, channelID *string) (*model.Monitor, error) {
	query := `UPDATE monitors SET is_active = $1, channel_id = $2, updated_at = now() WHERE id = $3
		RETURNING id, monitor_name, check_type, message, metadata, timeout, re_alert_interval, status, is_active, channel_id, server_ip, server_name, last_seen_at, created_at, updated_at`

	var m model.Monitor
	err := Pool.QueryRow(ctx, query, isActive, channelID, id).Scan(
		&m.ID, &m.MonitorName, &m.CheckType, &m.Message, &m.Metadata,
		&m.Timeout, &m.ReAlertInterval, &m.Status, &m.IsActive, &m.ChannelID,
		&m.ServerIP, &m.ServerName, &m.LastSeenAt, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func DeleteMonitor(ctx context.Context, id string) error {
	_, err := Pool.Exec(ctx, `DELETE FROM monitors WHERE id = $1`, id)
	return err
}

func SetMonitorStatus(ctx context.Context, id string, status string) error {
	_, err := Pool.Exec(ctx, `UPDATE monitors SET status = $1, updated_at = now() WHERE id = $2`, status, id)
	return err
}

// --- Active monitors that are overdue ---

type OverdueMonitor struct {
	model.Monitor
	TelegramChatID string
}

func GetOverdueMonitors(ctx context.Context) ([]OverdueMonitor, error) {
	query := `
		SELECT m.id, m.monitor_name, m.check_type, m.message, m.metadata,
			m.timeout, m.re_alert_interval, m.status, m.is_active, m.channel_id,
			m.server_ip, m.server_name, m.last_seen_at, m.created_at, m.updated_at,
			c.telegram_chat_id
		FROM monitors m
		JOIN notification_channels c ON m.channel_id = c.id
		WHERE m.is_active = true
		  AND m.last_seen_at + (m.timeout || ' seconds')::interval < now()`

	rows, err := Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []OverdueMonitor
	for rows.Next() {
		var om OverdueMonitor
		if err := rows.Scan(
			&om.ID, &om.MonitorName, &om.CheckType, &om.Message, &om.Metadata,
			&om.Timeout, &om.ReAlertInterval, &om.Status, &om.IsActive, &om.ChannelID,
			&om.ServerIP, &om.ServerName, &om.LastSeenAt, &om.CreatedAt, &om.UpdatedAt,
			&om.TelegramChatID,
		); err != nil {
			return nil, err
		}
		result = append(result, om)
	}
	return result, nil
}

// --- Recovered monitors (were down, now back up) ---

func GetRecoveredMonitors(ctx context.Context) ([]OverdueMonitor, error) {
	query := `
		SELECT m.id, m.monitor_name, m.check_type, m.message, m.metadata,
			m.timeout, m.re_alert_interval, m.status, m.is_active, m.channel_id,
			m.server_ip, m.server_name, m.last_seen_at, m.created_at, m.updated_at,
			c.telegram_chat_id
		FROM monitors m
		JOIN notification_channels c ON m.channel_id = c.id
		JOIN alert_states a ON a.monitor_id = m.id AND a.status = 'firing'
		WHERE m.is_active = true
		  AND m.status = 'up'`

	rows, err := Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []OverdueMonitor
	for rows.Next() {
		var om OverdueMonitor
		if err := rows.Scan(
			&om.ID, &om.MonitorName, &om.CheckType, &om.Message, &om.Metadata,
			&om.Timeout, &om.ReAlertInterval, &om.Status, &om.IsActive, &om.ChannelID,
			&om.ServerIP, &om.ServerName, &om.LastSeenAt, &om.CreatedAt, &om.UpdatedAt,
			&om.TelegramChatID,
		); err != nil {
			return nil, err
		}
		result = append(result, om)
	}
	return result, nil
}

// --- Alert States ---

func GetFiringAlert(ctx context.Context, monitorID string) (*model.AlertState, error) {
	query := `SELECT id, monitor_id, status, last_alerted_at, fired_at, resolved_at FROM alert_states WHERE monitor_id = $1 AND status = 'firing'`

	var a model.AlertState
	err := Pool.QueryRow(ctx, query, monitorID).Scan(&a.ID, &a.MonitorID, &a.Status, &a.LastAlertedAt, &a.FiredAt, &a.ResolvedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func CreateAlertState(ctx context.Context, monitorID string) error {
	_, err := Pool.Exec(ctx,
		`INSERT INTO alert_states (monitor_id, status, last_alerted_at, fired_at) VALUES ($1, 'firing', now(), now())`,
		monitorID)
	return err
}

func UpdateAlertLastAlerted(ctx context.Context, alertID string) error {
	_, err := Pool.Exec(ctx,
		`UPDATE alert_states SET last_alerted_at = now() WHERE id = $1`, alertID)
	return err
}

func ResolveAlert(ctx context.Context, alertID string) error {
	_, err := Pool.Exec(ctx,
		`UPDATE alert_states SET status = 'resolved', resolved_at = now() WHERE id = $1`, alertID)
	return err
}

// --- Notification Channels ---

func CreateChannel(ctx context.Context, name, telegramChatID string) (*model.NotificationChannel, error) {
	query := `INSERT INTO notification_channels (name, telegram_chat_id) VALUES ($1, $2) RETURNING id, name, telegram_chat_id, created_at`

	var ch model.NotificationChannel
	err := Pool.QueryRow(ctx, query, name, telegramChatID).Scan(&ch.ID, &ch.Name, &ch.TelegramChatID, &ch.CreatedAt)
	return &ch, err
}

func GetAllChannels(ctx context.Context) ([]model.NotificationChannel, error) {
	rows, err := Pool.Query(ctx, `SELECT id, name, telegram_chat_id, created_at FROM notification_channels ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []model.NotificationChannel
	for rows.Next() {
		var ch model.NotificationChannel
		if err := rows.Scan(&ch.ID, &ch.Name, &ch.TelegramChatID, &ch.CreatedAt); err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}
	return channels, nil
}

func DeleteChannel(ctx context.Context, id string) error {
	_, err := Pool.Exec(ctx, `DELETE FROM notification_channels WHERE id = $1`, id)
	return err
}

// --- Notification Logs ---

func CreateNotificationLog(ctx context.Context, monitorID string, channelID *string, alertType, message string, success bool, errMsg string) error {
	_, err := Pool.Exec(ctx,
		`INSERT INTO notification_logs (monitor_id, channel_id, alert_type, message, success, error) VALUES ($1, $2, $3, $4, $5, $6)`,
		monitorID, channelID, alertType, message, success, errMsg)
	return err
}

func GetNotificationLogs(ctx context.Context) ([]model.NotificationLog, error) {
	rows, err := Pool.Query(ctx,
		`SELECT id, monitor_id, channel_id, alert_type, message, success, error, created_at FROM notification_logs ORDER BY created_at DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []model.NotificationLog
	for rows.Next() {
		var l model.NotificationLog
		if err := rows.Scan(&l.ID, &l.MonitorID, &l.ChannelID, &l.AlertType, &l.Message, &l.Success, &l.Error, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

// --- Helpers ---

func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}
