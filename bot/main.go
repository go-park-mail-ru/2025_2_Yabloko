package main

import (
	"log"

	bot "bot/internal"
)

func main() {
	cfg := bot.LoadConfig()

	b, err := bot.NewBot(cfg)
	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}

	b.Run()
}
