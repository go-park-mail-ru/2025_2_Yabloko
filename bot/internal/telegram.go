package internal

import (
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	cfg Config
	api *tgbotapi.BotAPI
}

func NewBot(cfg Config) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		return nil, err
	}
	return &Bot{cfg: cfg, api: api}, nil
}

func (b *Bot) Run() {
	log.Printf("Authorized on account %s", b.api.Self.UserName)
	log.Printf("Allowed chat IDs: %v", b.cfg.ChatIDs)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		msg := update.Message

		if !msg.Chat.IsPrivate() && !isAllowedChat(msg.Chat.ID, b.cfg.ChatIDs) {
			continue
		}

		if !isMentioned(b.api.Self.UserName, msg) {
			continue
		}

		text := extractTextAfterMention(b.api.Self.UserName, msg.Text)
		if strings.TrimSpace(text) == "" {
			text = "скажи что-нибудь нормальное"
		}

		reply := GenerateReply(text, b.cfg)

		resp := tgbotapi.NewMessage(msg.Chat.ID, reply)
		resp.ReplyToMessageID = msg.MessageID

		if _, err := b.api.Send(resp); err != nil {
			log.Printf("failed to send message: %v", err)
		}
	}
}

func isAllowedChat(chatID int64, allowed []int64) bool {
	for _, id := range allowed {
		if id == chatID {
			return true
		}
	}
	return false
}

func isMentioned(botUsername string, msg *tgbotapi.Message) bool {
	if msg.Chat.IsPrivate() {
		return true
	}

	if len(msg.Entities) > 0 {
		for _, e := range msg.Entities {
			if e.Type == "mention" {
				mention := msg.Text[e.Offset : e.Offset+e.Length]
				if mention == "@"+botUsername {
					return true
				}
			}
		}
	}

	return strings.Contains(msg.Text, "@"+botUsername)
}

func extractTextAfterMention(botUsername, text string) string {
	idx := strings.Index(text, "@"+botUsername)
	if idx == -1 {
		return text
	}
	after := text[idx+len(botUsername)+1:]
	return strings.TrimSpace(after)
}
