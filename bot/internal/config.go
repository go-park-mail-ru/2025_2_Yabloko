package internal

import (
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	TelegramToken string
	ChatIDs       []int64
	HFToken       string
}

func LoadConfig() Config {
	token := os.Getenv("ALERT_TELEGRAM_BOT_TOKEN")
	chatIDStr := os.Getenv("ALLOW_CHAT_IDS")
	if token == "" || chatIDStr == "" {
		log.Fatal("ALERT_TELEGRAM_BOT_TOKEN or ALLOW_CHAT_IDS not set")
	}

	chatIDs := parseChatIDs(chatIDStr)
	if len(chatIDs) == 0 {
		log.Fatal("no valid chat IDs provided")
	}

	hfToken := os.Getenv("HF_API_TOKEN")
	if hfToken == "" {
		log.Println("HF_API_TOKEN is empty, LLM replies will be disabled")
	}

	return Config{
		TelegramToken: token,
		ChatIDs:       chatIDs,
		HFToken:       hfToken,
	}
}

func parseChatIDs(raw string) []int64 {
	parts := strings.Split(raw, ",")
	var ids []int64
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			log.Printf("invalid chat ID '%s', skipping: %v", p, err)
			continue
		}
		ids = append(ids, id)
	}
	return ids
}
