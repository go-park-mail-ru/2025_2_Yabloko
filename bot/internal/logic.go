package internal

import (
	"math/rand/v2"
	"strings"
)

func GenerateReply(input string, cfg Config) string {
	lower := strings.ToLower(input)

	badWords := []string{
		"иди нахуй",
		"пошел нахуй",
		"иди на хуй",
		"нахуй",
	}

	for _, w := range badWords {
		if strings.Contains(lower, w) {
			return "сам иди нахуй"
		}
	}

	if strings.Contains(lower, "привет") {
		return "Привет, заебал тэгать, чё надо?"
	}

	if strings.Contains(lower, "кто ты") || strings.Contains(lower, "что ты") {
		return "Я бот для рофлов, не еби мозгу"
	}

	// 5% — рофл
	if rand.Float64() < 0.05 {
		return randomRoast()
	}

	if cfg.HFToken != "" {
		if llmAns, err := CallLLM(cfg.HFToken, input); err == nil && strings.TrimSpace(llmAns) != "" {
			return llmAns
		}
	}

	return randomRoast()
}

func randomRoast() string {
	answers := []string{
		"Я тебя услышал, но мне похуй",
		"Разберись сам",
		"Завязывай с этим",
		"Пошёл код писать, скотина",
		"Сделай ПР и не ной",
		"Услышал тебя, дорогой",
		"Услышал тебя, родной",
		"ГАААЗ Дядь",
		"Газуй",
		"Газуем",
	}
	return answers[rand.IntN(len(answers))]
}
