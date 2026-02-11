package watcher

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/mohsen/alertingGo/db"
	"github.com/mohsen/alertingGo/notifier"
)

func Start() {
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for range ticker.C {
			checkOverdue()
			checkRecovered()
		}
	}()
	log.Println("watcher started (every 10s)")
}

func checkOverdue() {
	ctx := context.Background()

	monitors, err := db.GetOverdueMonitors(ctx)
	if err != nil {
		log.Printf("[watcher] error fetching overdue monitors: %v", err)
		return
	}

	for _, om := range monitors {
		// Mark monitor as down
		if om.Status != "down" {
			if err := db.SetMonitorStatus(ctx, om.ID, "down"); err != nil {
				log.Printf("[watcher] error setting monitor %s to down: %v", om.ID, err)
				continue
			}
		}

		// Check existing alert state
		alert, err := db.GetFiringAlert(ctx, om.ID)
		if err != nil && err != pgx.ErrNoRows {
			log.Printf("[watcher] error fetching alert for monitor %s: %v", om.ID, err)
			continue
		}

		downSince := time.Since(om.LastSeenAt)

		if alert == nil {
			// First alert — create alert state and fire
			if err := db.CreateAlertState(ctx, om.ID); err != nil {
				log.Printf("[watcher] error creating alert state for %s: %v", om.ID, err)
				continue
			}

			msg := fmt.Sprintf("🔴 *ALERT: %s (%s) is DOWN*\nLast seen: %s ago\nTimeout: %ds\nMessage: %s",
				om.MonitorName, om.CheckType, db.FormatDuration(downSince), om.Timeout, om.Message)
			success, errMsg := notifier.SendTelegram(om.TelegramChatID, msg)
			db.CreateNotificationLog(ctx, om.ID, om.ChannelID, "alert", msg, success, errMsg)

		} else {
			// Re-alert if re_alert_interval has passed
			sinceLast := time.Since(alert.LastAlertedAt)
			if sinceLast >= time.Duration(om.ReAlertInterval)*time.Second {
				if err := db.UpdateAlertLastAlerted(ctx, alert.ID); err != nil {
					log.Printf("[watcher] error updating alert %s: %v", alert.ID, err)
					continue
				}

				msg := fmt.Sprintf("🔴 *RE-ALERT: %s (%s) still DOWN*\nDown for: %s\nMessage: %s",
					om.MonitorName, om.CheckType, db.FormatDuration(downSince), om.Message)
				success, errMsg := notifier.SendTelegram(om.TelegramChatID, msg)
				db.CreateNotificationLog(ctx, om.ID, om.ChannelID, "re_alert", msg, success, errMsg)
			}
		}
	}
}

func checkRecovered() {
	ctx := context.Background()

	monitors, err := db.GetRecoveredMonitors(ctx)
	if err != nil {
		log.Printf("[watcher] error fetching recovered monitors: %v", err)
		return
	}

	for _, om := range monitors {
		alert, err := db.GetFiringAlert(ctx, om.ID)
		if err != nil {
			log.Printf("[watcher] error fetching alert for recovered monitor %s: %v", om.ID, err)
			continue
		}

		downtime := time.Since(alert.FiredAt)

		if err := db.ResolveAlert(ctx, alert.ID); err != nil {
			log.Printf("[watcher] error resolving alert %s: %v", alert.ID, err)
			continue
		}

		msg := fmt.Sprintf("🟢 *RECOVERED: %s (%s) is back UP*\nWas down for: %s",
			om.MonitorName, om.CheckType, db.FormatDuration(downtime))
		success, errMsg := notifier.SendTelegram(om.TelegramChatID, msg)
		db.CreateNotificationLog(ctx, om.ID, om.ChannelID, "recovered", msg, success, errMsg)
	}
}
