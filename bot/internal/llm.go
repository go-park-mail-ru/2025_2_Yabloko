package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type hfMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type hfRequest struct {
	Model    string      `json:"model"`
	Messages []hfMessage `json:"messages"`
}

type hfChoice struct {
	Message hfMessage `json:"message"`
}

type hfResponse struct {
	Choices []hfChoice `json:"choices"`
}

func CallLLM(hfToken, prompt string) (string, error) {
	if hfToken == "" {
		log.Println("[hf] no HF_API_TOKEN")
		return "", fmt.Errorf("no HF_API_TOKEN")
	}

	reqBody := hfRequest{
		Model: "swiss-ai/Apertus-70B-Instruct-2509",
		Messages: []hfMessage{
			{Role: "user", Content: prompt},
		},
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("[hf] marshal error: %v", err)
		return "", err
	}
	log.Printf("[hf] request body: %s", string(data))

	client := &http.Client{Timeout: 20 * time.Second}

	req, err := http.NewRequest("POST", "https://router.huggingface.co/v1/chat/completions", bytes.NewReader(data))
	if err != nil {
		log.Printf("[hf] new request error: %v", err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+hfToken)

	log.Println("[hf] sending request...")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[hf] http error: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[hf] read body error: %v", err)
		return "", err
	}
	log.Printf("[hf] status: %d, raw: %s", resp.StatusCode, string(bodyBytes))

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("hf status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var out hfResponse
	if err := json.Unmarshal(bodyBytes, &out); err != nil {
		log.Printf("[hf] decode json error: %v", err)
		return "", err
	}
	if len(out.Choices) == 0 {
		log.Println("[hf] empty choices")
		return "", nil
	}

	return strings.TrimSpace(out.Choices[0].Message.Content), nil
}
