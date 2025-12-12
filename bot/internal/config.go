package internal

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	TelegramToken string
	ChatID        int64
	HFToken       string
}

func LoadConfig() Config {
	token := os.Getenv("ALERT_TELEGRAM_BOT_TOKEN")
	chatIDStr := os.Getenv("ALERT_TELEGRAM_CHAT_ID")
	if token == "" || chatIDStr == "" {
		log.Fatal("ALERT_TELEGRAM_BOT_TOKEN or ALERT_TELEGRAM_CHAT_ID not set")
	}

	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		log.Fatalf("invalid ALERT_TELEGRAM_CHAT_ID: %v", err)
	}

	hfToken := os.Getenv("HF_API_TOKEN")
	if hfToken == "" {
		log.Println("HF_API_TOKEN is empty, LLM replies will be disabled")
	}

	return Config{
		TelegramToken: token,
		ChatID:        chatID,
		HFToken:       hfToken,
	}
}
