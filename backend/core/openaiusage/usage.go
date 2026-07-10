package openaiusage

import (
	"encoding/json"
	"strconv"
	"strings"

	"voice-chat-api-go/logger"
)

// Usage is the top-level "usage" object from OpenAI chat/completions responses.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ParseFromChatCompletion extracts usage from a chat/completions JSON body.
func ParseFromChatCompletion(respBytes []byte) (Usage, bool) {
	var wrap struct {
		Usage *Usage `json:"usage"`
	}
	if err := json.Unmarshal(respBytes, &wrap); err != nil || wrap.Usage == nil {
		return Usage{}, false
	}
	u := *wrap.Usage
	if u.TotalTokens <= 0 && u.PromptTokens+u.CompletionTokens > 0 {
		u.TotalTokens = u.PromptTokens + u.CompletionTokens
	}
	return u, true
}

// UsagePresent is true when the completion response included non-zero usage fields.
func UsagePresent(u Usage) bool {
	return u.PromptTokens > 0 || u.CompletionTokens > 0 || u.TotalTokens > 0
}

// LogTok logs token usage for one OpenAI completion. component is chat | world_interpret | motion | long_memory_extract.
// httpPath may be empty. turnID is optional (e.g. X-Turn-Id for correlating multiple HTTP calls client-side).
// systemPromptTokensApprox / userContentTokensApprox が 0 以上なら tiktoken 近似を付与（messages の content 相当のみ）。
func LogTok(component, httpPath, turnID string, u Usage, ok bool, systemPromptTokensApprox, userContentTokensApprox int) {
	if !ok {
		return
	}
	var b strings.Builder
	b.WriteString("openai_tokens component=")
	b.WriteString(component)
	b.WriteString(" prompt_tokens=")
	b.WriteString(strconv.Itoa(u.PromptTokens))
	b.WriteString(" completion_tokens=")
	b.WriteString(strconv.Itoa(u.CompletionTokens))
	b.WriteString(" total_tokens=")
	b.WriteString(strconv.Itoa(u.TotalTokens))
	if systemPromptTokensApprox >= 0 {
		b.WriteString(" system_prompt_tokens≈")
		b.WriteString(strconv.Itoa(systemPromptTokensApprox))
	}
	if userContentTokensApprox >= 0 {
		b.WriteString(" user_content_tokens≈")
		b.WriteString(strconv.Itoa(userContentTokensApprox))
	}
	if httpPath != "" {
		b.WriteString(" http_path=")
		b.WriteString(httpPath)
	}
	if turnID != "" {
		b.WriteString(" turn_id=")
		b.WriteString(turnID)
	}
	logger.Info(b.String())
}

// LogHTTPRequestTotals logs OpenAI token totals for this HTTP request (one completion → same numbers as LogTok).
// Use together with LogTok: component別と「このリクエスト全体」を両方出す。
func LogHTTPRequestTotals(httpPath, turnID string, u Usage, ok bool, systemPromptTokensApprox, userContentTokensApprox int) {
	if !ok {
		return
	}
	var b strings.Builder
	b.WriteString("openai_http_request_totals http_path=")
	b.WriteString(httpPath)
	b.WriteString(" openai_calls=1 prompt_tokens=")
	b.WriteString(strconv.Itoa(u.PromptTokens))
	b.WriteString(" completion_tokens=")
	b.WriteString(strconv.Itoa(u.CompletionTokens))
	b.WriteString(" total_tokens=")
	b.WriteString(strconv.Itoa(u.TotalTokens))
	if systemPromptTokensApprox >= 0 {
		b.WriteString(" system_prompt_tokens≈")
		b.WriteString(strconv.Itoa(systemPromptTokensApprox))
	}
	if userContentTokensApprox >= 0 {
		b.WriteString(" user_content_tokens≈")
		b.WriteString(strconv.Itoa(userContentTokensApprox))
	}
	if turnID != "" {
		b.WriteString(" turn_id=")
		b.WriteString(turnID)
	}
	logger.Info(b.String())
}

// LogForOpenAIHandler logs per-component usage and this HTTP request's OpenAI totals (one completion).
// systemPrompt / userContent は送信に使った各 role の全文（空なら対応する ≈ フィールドは出さない）。
func LogForOpenAIHandler(component, httpPath, turnID string, u Usage, ok bool, systemPrompt, userContent string) {
	enc := resolveEncodingModel(component)
	sysApprox := -1
	if strings.TrimSpace(systemPrompt) != "" {
		sysApprox = ApproxTextTokens(enc, systemPrompt)
	}
	userApprox := -1
	if strings.TrimSpace(userContent) != "" {
		userApprox = ApproxTextTokens(enc, userContent)
	}
	LogTok(component, httpPath, turnID, u, ok, sysApprox, userApprox)
	LogHTTPRequestTotals(httpPath, turnID, u, ok, sysApprox, userApprox)
}
