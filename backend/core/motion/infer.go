package motion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"voice-chat-api-go/core/openaiusage"
	"voice-chat-api-go/logger"
)

// Infer calls OpenAI chat completions with structured output (motion script only).
func Infer(systemPrompt, userContent string) (*ScriptOut, openaiusage.Usage, error) {
	systemPrompt = strings.TrimSpace(systemPrompt)
	userContent = strings.TrimSpace(userContent)
	if systemPrompt == "" {
		return nil, openaiusage.Usage{}, fmt.Errorf("empty system prompt")
	}
	if userContent == "" {
		return nil, openaiusage.Usage{}, fmt.Errorf("empty user content")
	}

	model := strings.TrimSpace(os.Getenv("MOTION_MODEL"))
	if model == "" {
		model = strings.TrimSpace(os.Getenv("CHAT_MODEL"))
	}
	if model == "" {
		model = "gpt-4.1-nano"
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, openaiusage.Usage{}, fmt.Errorf("OPENAI_API_KEY is not set")
	}

	msgs := []map[string]string{
		{"role": "system", "content": systemPrompt},
		{"role": "user", "content": userContent},
	}

	reqBody := map[string]any{
		"model":    model,
		"messages": msgs,
		"stream":   false,
		"response_format": map[string]any{
			"type": "json_schema",
			"json_schema": map[string]any{
				"name":   "avatar_motion_script",
				"strict": true,
				"schema": MotionScriptJSONSchema(),
			},
		},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, openaiusage.Usage{}, fmt.Errorf("encode request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, openaiusage.Usage{}, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	t0 := time.Now()
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, openaiusage.Usage{}, fmt.Errorf("openai request: %w", err)
	}
	defer resp.Body.Close()
	respBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, openaiusage.Usage{}, fmt.Errorf("openai read body: %w", readErr)
	}
	if resp.StatusCode != http.StatusOK {
		logger.Warn("motion infer: openai status=%d body=%s", resp.StatusCode, truncateForLog(respBytes, 500))
		return nil, openaiusage.Usage{}, fmt.Errorf("openai status %d", resp.StatusCode)
	}
	usage, _ := openaiusage.ParseFromChatCompletion(respBytes)
	logger.Info("motion infer: ok duration=%s model=%s", time.Since(t0).Round(time.Millisecond), model)

	var openaiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Refusal *string `json:"refusal,omitempty"`
	}
	if err := json.Unmarshal(respBytes, &openaiResp); err != nil {
		return nil, usage, fmt.Errorf("decode openai response: %w", err)
	}
	if openaiResp.Refusal != nil && *openaiResp.Refusal != "" {
		return nil, usage, fmt.Errorf("model refused: %s", *openaiResp.Refusal)
	}
	content := ""
	if len(openaiResp.Choices) > 0 {
		content = openaiResp.Choices[0].Message.Content
	}
	var out ScriptOut
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		logger.Warn("motion infer: parse JSON: %v raw=%s", err, truncateForLog([]byte(content), 400))
		return nil, usage, fmt.Errorf("parse motion script: %w", err)
	}
	if len(out.Lines) == 0 {
		return nil, usage, fmt.Errorf("empty lines")
	}
	return &out, usage, nil
}

func truncateForLog(b []byte, max int) string {
	s := string(b)
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
