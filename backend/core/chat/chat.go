package chat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"voice-chat-api-go/config"
	"voice-chat-api-go/core/openaiusage"
	"voice-chat-api-go/logger"

	"github.com/pkoukk/tiktoken-go"
)

// Chat メッセージを渡して OpenAI で応答を生成する。systemPrompt は呼び出し側で読み込み・前処理して渡す。
// Structured Outputs (json_schema) で reply + style_local_id の JSON を取得し、TTS 用 lines は SegmentReplyToAvatarScript のあと FilterSpeakableAvatarScript で正規化する。
// requestSpacePersonaID は HTTP の space_persona_id（TTS・履歴の話者 id。script JSON には含めない）。
// speakerDisplayName は HTTP 話者の表示名（空可）。participantDisplayLabels は小文字の space_persona_id → 表示名（履歴用・nil 可）。
func Chat(messages []Message, systemPrompt string, requestSpacePersonaID string, speakerDisplayName string, participantDisplayLabels map[string]string) (*Result, openaiusage.Usage, error) {
	messagesWithSystem := append([]Message{
		{Role: "system", Content: systemPrompt},
	}, messages...)

	chatModel := os.Getenv("CHAT_MODEL")
	if chatModel == "" {
		chatModel = "gpt-4o-mini" // Structured Outputs 対応モデル
	}

	maxTokBudget := config.ChatTrimMaxTokens()
	trimmed := trimMessagesWithSystem(messagesWithSystem, maxTokBudget, chatModel)
	reqSpeakerID := strings.TrimSpace(requestSpacePersonaID)
	messagesWithoutSystem := trimmed[1:]

	logChatOpenAIRequest(chatModel, trimmed, maxTokBudget, reqSpeakerID)

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, openaiusage.Usage{}, fmt.Errorf("OPENAI_API_KEY is not set")
	}

	reqBody := map[string]any{
		"model":    chatModel,
		"messages": trimmed,
		"stream":   false,
		"response_format": map[string]any{
			"type": "json_schema",
			"json_schema": map[string]any{
				"name":   "chat_reply",
				"strict": true,
				"schema": ChatReplyJSONSchema(),
			},
		},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, openaiusage.Usage{}, fmt.Errorf("encode request: %w", err)
	}
	logger.Info("chat request: http_body_bytes=%d (JSON に含まれる messages 全文)", len(body))

	httpReq, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, openaiusage.Usage{}, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	openAIT0 := time.Now()
	resp, err := client.Do(httpReq)
	if err != nil {
		logger.Info("openai chat/completions: failed after %s model=%s err=%v", time.Since(openAIT0).Round(time.Millisecond), chatModel, err)
		return nil, openaiusage.Usage{}, fmt.Errorf("openai request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, readErr := io.ReadAll(resp.Body)
	openAIDuration := time.Since(openAIT0)
	if readErr != nil {
		logger.Info("openai chat/completions: read body error after %s model=%s err=%v", openAIDuration.Round(time.Millisecond), chatModel, readErr)
		return nil, openaiusage.Usage{}, fmt.Errorf("openai read body: %w", readErr)
	}
	if resp.StatusCode != http.StatusOK {
		logger.Info("openai chat/completions: status=%d duration=%s model=%s responseBytes=%d", resp.StatusCode, openAIDuration.Round(time.Millisecond), chatModel, len(respBytes))
		return nil, openaiusage.Usage{}, fmt.Errorf("openai status %d: %s", resp.StatusCode, string(respBytes))
	}
	usage, _ := openaiusage.ParseFromChatCompletion(respBytes)
	logger.Info("openai chat/completions: ok duration=%s model=%s responseBytes=%d request_messages=%d", openAIDuration.Round(time.Millisecond), chatModel, len(respBytes), len(trimmed))

	var openaiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Refusal *string `json:"refusal,omitempty"`
	}
	if err := json.Unmarshal(respBytes, &openaiResp); err != nil {
		return nil, usage, fmt.Errorf("decode response: %w", err)
	}

	if openaiResp.Refusal != nil && *openaiResp.Refusal != "" {
		return nil, usage, fmt.Errorf("model refused: %s", *openaiResp.Refusal)
	}

	content := ""
	if len(openaiResp.Choices) > 0 {
		content = openaiResp.Choices[0].Message.Content
	}

	var llmOut struct {
		Reply         string `json:"reply"`
		StyleLocalID  string `json:"style_local_id"`
	}
	if err := json.Unmarshal([]byte(content), &llmOut); err != nil {
		logger.Error("chat reply parse error: %v, raw=%s", err, content)
		return nil, usage, fmt.Errorf("parse chat reply: %w", err)
	}
	reply := strings.TrimSpace(llmOut.Reply)
	if reply == "" {
		return nil, usage, fmt.Errorf("parse chat reply: empty reply")
	}
	logChatReplyOpenAIContent(content)
	script := FilterSpeakableAvatarScript(SegmentReplyToAvatarScript(reply, llmOut.StyleLocalID))
	if len(script.Lines) == 0 {
		return nil, usage, fmt.Errorf("segment reply: empty lines")
	}
	logChatReplySegmented(script, reply)

	displayContent := avatarScriptToDisplayText(script)

	assistantContent := displayContent
	if j, err := MarshalAssistantHistoryFromScript(&script, participantDisplayLabels, reqSpeakerID, speakerDisplayName); err == nil {
		assistantContent = j
	} else {
		logger.Warn("MarshalAssistantHistoryFromScript: %v", err)
		firstSp := reqSpeakerID
		fb, err2 := json.Marshal([]AssistantHistoryItem{{ID: firstSp, Name: strings.TrimSpace(speakerDisplayName), Message: displayContent}})
		if err2 == nil {
			assistantContent = string(fb)
		}
	}

	return &Result{
		Histories: append(messagesWithoutSystem, Message{Role: "assistant", Content: assistantContent, CreatedAt: time.Now()}),
		Content:   displayContent,
		Script:    &script,
	}, usage, nil
}

func displayLabelForSpeaker(labels map[string]string, speakerID string) string {
	if labels == nil {
		return ""
	}
	k := strings.ToLower(strings.TrimSpace(speakerID))
	if k == "" {
		return ""
	}
	return strings.TrimSpace(labels[k])
}

// MarshalAssistantHistoryFromScript assistant 履歴用 JSON 配列 [{id,name,message}, ...]。
func MarshalAssistantHistoryFromScript(script *AvatarScript, labels map[string]string, httpSpeakerID, httpSpeakerDisplayName string) (string, error) {
	if script == nil || len(script.Lines) == 0 {
		return "", fmt.Errorf("empty script")
	}
	msg := strings.TrimSpace(avatarScriptToDisplayText(*script))
	if msg == "" {
		return "", fmt.Errorf("no assistant lines")
	}
	httpSID := strings.TrimSpace(httpSpeakerID)
	httpName := strings.TrimSpace(httpSpeakerDisplayName)
	name := displayLabelForSpeaker(labels, httpSID)
	if name == "" {
		name = httpName
	}
	if name == "" {
		name = "アシスタント"
	}
	b, err := json.Marshal([]AssistantHistoryItem{{ID: httpSID, Name: name, Message: msg}})
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func avatarScriptToDisplayText(script AvatarScript) string {
	parts := make([]string, 0, len(script.Lines))
	for _, line := range script.Lines {
		if t := strings.TrimSpace(line.Text); t != "" {
			parts = append(parts, t)
		}
	}
	return strings.Join(parts, "\n")
}

func trimMessagesWithSystem(messages []Message, maxTokens int, model string) []Message {
	if len(messages) == 0 {
		return messages
	}
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	enc, err := tiktoken.EncodingForModel(model)
	if err != nil {
		enc, err = tiktoken.EncodingForModel("gpt-4")
		if err != nil {
			panic(err)
		}
	}

	systemMsg := messages[0]
	otherMsgs := messages[1:]

	systemTokens := len(enc.Encode(systemMsg.Content, nil, nil))
	if systemTokens > maxTokens {
		logger.Error(
			"trimMessagesWithSystem: system prompt exceeds maxTokens (would return empty → panic risk in caller): model=%s systemTokens=%d maxTokens=%d systemChars=%d",
			model, systemTokens, maxTokens, len(systemMsg.Content),
		)
		return []Message{}
	}

	totalTokens := systemTokens
	tokenCounts := make([]int, len(otherMsgs))
	for i, m := range otherMsgs {
		t := len(enc.Encode(m.Content, nil, nil))
		tokenCounts[i] = t
		totalTokens += t
	}

	startIdx := 0
	for totalTokens > maxTokens && startIdx < len(otherMsgs) {
		totalTokens -= tokenCounts[startIdx]
		startIdx++
	}
	if startIdx > 0 {
		logger.Warn(
			"trimMessagesWithSystem: dropped oldest user messages: model=%s startIdx=%d droppedMsgs=%d finalTotalTokens=%d maxTokens=%d systemTokens=%d",
			model, startIdx, startIdx, totalTokens, maxTokens, systemTokens,
		)
	}

	return append([]Message{systemMsg}, otherMsgs[startIdx:]...)
}

// logChatOpenAIRequest 送信直前の要約をターミナルに出す。
// content のトークン数は tiktoken（role 名や API オーバーヘッドは含まない近似）。
func logChatOpenAIRequest(model string, trimmed []Message, maxTokBudget int, speakerSpacePersonaID string) {
	if len(trimmed) == 0 {
		logger.Info("chat request: (no messages after trim) model=%s", model)
		return
	}
	enc, err := tiktoken.EncodingForModel(model)
	if err != nil {
		enc, err = tiktoken.EncodingForModel("gpt-4")
		if err != nil {
			logger.Warn("chat request log: tiktoken unavailable: %v", err)
			return
		}
	}

	sysTok := len(enc.Encode(trimmed[0].Content, nil, nil))
	histTok := 0
	for i := 1; i < len(trimmed); i++ {
		histTok += len(enc.Encode(trimmed[i].Content, nil, nil))
	}
	total := sysTok + histTok

	logger.Info(
		"chat request → openai: model=%s CHAT_TRIM_MAX_TOKENS=%d messages=%d content_tokens≈%d (system=%d history=%d) space_persona_id=%q",
		model, maxTokBudget, len(trimmed), total, sysTok, histTok, speakerSpacePersonaID,
	)
}

// CleanSystemPrompt システムプロンプトの空白を除去する（呼び出し側で渡す前に使う）
func CleanSystemPrompt(raw string) string {
	re := regexp.MustCompile(`[\s　]+`)
	return re.ReplaceAllString(raw, "")
}

const chatReplyLogMaxRunes = 1200

func truncateRunesForLog(s string, maxRunes int) string {
	if maxRunes < 1 {
		return ""
	}
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	rs := []rune(s)
	if maxRunes <= 1 {
		return "…"
	}
	return string(rs[:maxRunes-1]) + "…"
}

// logChatReplyOpenAIContent OpenAI choices[0].message.content（JSON 文字列）を INFO に出す。
func logChatReplyOpenAIContent(raw string) {
	trunc := truncateRunesForLog(raw, chatReplyLogMaxRunes)
	truncNote := ""
	if trunc != raw {
		truncNote = " (truncated for log)"
	}
	logger.Info("chat_reply openai message.content runes=%d%s: %s", utf8.RuneCountInString(raw), truncNote, trunc)
}

// logChatReplySegmented パース後の style とサーバ分割結果の要約。
func logChatReplySegmented(script AvatarScript, reply string) {
	previews := make([]string, 0, len(script.Lines))
	for i, l := range script.Lines {
		t := strings.TrimSpace(l.Text)
		previews = append(previews, fmt.Sprintf("%d:%s", i+1, truncateRunesForLog(t, 36)))
	}
	logger.Info(
		"chat_reply segmented: style_local_id=%q reply_runes=%d lines=%d | %s",
		script.StyleLocalID,
		utf8.RuneCountInString(reply),
		len(script.Lines),
		strings.Join(previews, " | "),
	)
}
