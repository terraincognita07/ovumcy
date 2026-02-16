package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/terraincognita07/lume/internal/models"
	"gorm.io/gorm"
)

type NotificationService struct {
	db                     *gorm.DB
	botToken               string
	chatID                 string
	enabled                bool
	periodReminderDays     int
	fertilityReminder      bool
	location               *time.Location
	client                 *http.Client
	mu                     sync.Mutex
	sentDailyNotifications map[string]time.Time
}

func NewNotificationService(db *gorm.DB, location *time.Location) *NotificationService {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	enabled := botToken != "" && chatID != ""

	periodReminderDays := 2
	if raw := os.Getenv("TELEGRAM_PERIOD_REMINDER_DAYS"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed >= 0 {
			periodReminderDays = parsed
		}
	}

	fertilityReminder := true
	if raw := os.Getenv("TELEGRAM_NOTIFY_FERTILITY"); raw != "" {
		fertilityReminder = raw == "1" || raw == "true" || raw == "TRUE"
	}

	if location == nil {
		location = time.Local
	}

	return &NotificationService{
		db:                 db,
		botToken:           botToken,
		chatID:             chatID,
		enabled:            enabled,
		periodReminderDays: periodReminderDays,
		fertilityReminder:  fertilityReminder,
		location:           location,
		client: &http.Client{
			Timeout: 8 * time.Second,
		},
		sentDailyNotifications: make(map[string]time.Time),
	}
}

func (service *NotificationService) Start(ctx context.Context) {
	if !service.enabled {
		return
	}

	ticker := time.NewTicker(6 * time.Hour)
	go func() {
		defer ticker.Stop()

		service.run(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				service.run(ctx)
			}
		}
	}()
}

func (service *NotificationService) run(ctx context.Context) {
	owners := make([]models.User, 0)
	if err := service.db.WithContext(ctx).
		Where("role = ?", models.RoleOwner).
		Find(&owners).Error; err != nil {
		log.Printf("notifications: fetch owners failed: %v", err)
		return
	}

	now := time.Now().In(service.location)
	today := dateOnly(now)

	for _, owner := range owners {
		logs := make([]models.DailyLog, 0)
		from := today.AddDate(0, 0, -420)

		if err := service.db.WithContext(ctx).
			Where("user_id = ? AND date >= ? AND date <= ?", owner.ID, from, today).
			Order("date ASC").
			Find(&logs).Error; err != nil {
			log.Printf("notifications: fetch logs failed for user %d: %v", owner.ID, err)
			continue
		}

		stats := BuildCycleStats(logs, now, 14)
		if stats.NextPeriodStart.IsZero() {
			continue
		}

		daysUntilPeriod := int(stats.NextPeriodStart.Sub(today).Hours() / 24)
		if daysUntilPeriod == service.periodReminderDays {
			key := fmt.Sprintf("period:%d:%s", owner.ID, today.Format("2006-01-02"))
			if service.shouldSend(key, today) {
				message := fmt.Sprintf("Lume reminder: your predicted period starts in %d day(s) on %s.",
					service.periodReminderDays,
					stats.NextPeriodStart.Format("Jan 2"),
				)
				if err := service.sendTelegram(ctx, message); err != nil {
					log.Printf("notifications: send period reminder failed: %v", err)
				}
			}
		}

		if service.fertilityReminder && sameDay(today, stats.FertilityWindowStart) {
			key := fmt.Sprintf("fertility:%d:%s", owner.ID, today.Format("2006-01-02"))
			if service.shouldSend(key, today) {
				message := fmt.Sprintf("Lume reminder: your fertility window starts today (%s).",
					stats.FertilityWindowStart.Format("Jan 2"),
				)
				if err := service.sendTelegram(ctx, message); err != nil {
					log.Printf("notifications: send fertility reminder failed: %v", err)
				}
			}
		}
	}
}

func (service *NotificationService) shouldSend(key string, today time.Time) bool {
	service.mu.Lock()
	defer service.mu.Unlock()

	if sentOn, ok := service.sentDailyNotifications[key]; ok && sameDay(sentOn, today) {
		return false
	}

	service.sentDailyNotifications[key] = today
	if len(service.sentDailyNotifications) > 500 {
		service.sentDailyNotifications = make(map[string]time.Time)
	}
	return true
}

func (service *NotificationService) sendTelegram(ctx context.Context, message string) error {
	values := url.Values{}
	values.Set("chat_id", service.chatID)
	values.Set("text", message)

	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", service.botToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := service.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("telegram status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
