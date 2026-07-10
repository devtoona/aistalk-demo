package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"voice-chat-api-go/config"
	"voice-chat-api-go/core/chat"
	"voice-chat-api-go/core/openaiusage"
	"voice-chat-api-go/logger"
	"voice-chat-api-go/prompts"
)

type ChatResponse struct {
	Histories []chat.Message     `json:"histories"`
	Content   string             `json:"content"`
	Script    *chat.AvatarScript `json:"script"`
}

type ChatParticipantPayload struct {
	SpacePersonaID string `json:"space_persona_id"`
	Label          string `json:"label,omitempty"`
	PersonaKind    string `json:"persona_kind,omitempty"`
}

type PersonaPayload struct {
	Name          string `json:"name"`
	Personality   string `json:"personality"`
	ResponseStyle string `json:"response_style"`
	Tempo         string `json:"tempo"`
}

type ChatRequest struct {
	Messages       []chat.ChatClientMessage `json:"messages"`
	SpacePersonaID string                   `json:"space_persona_id"`
	Participants   []ChatParticipantPayload `json:"participants,omitempty"`
	Persona        *PersonaPayload          `json:"persona,omitempty"`
}

type ChatHandler struct{}

func NewChatHandler() *ChatHandler {
	return &ChatHandler{}
}

func formatDialogueParticipantsRoster(req ChatRequest) string {
	if len(req.Participants) > 0 {
		speakerID := strings.TrimSpace(req.SpacePersonaID)
		var b strings.Builder
		b.WriteString("(話者以外の参加者。会話は 1 対 1（相手はプレイヤー想定）。マイクは常にプレイヤーへ戻る。)")
		b.WriteString("いま喋っているペルソナの ID は HTTP リクエストの space_persona_id であり、この一覧からは除外している。)\n")
		for _, p := range req.Participants {
			id := strings.TrimSpace(p.SpacePersonaID)
			if id == "" || id == speakerID {
				continue
			}
			b.WriteString("space_persona_id: ")
			b.WriteString(id)
			if label := strings.TrimSpace(p.Label); label != "" {
				b.WriteString(" — 表示名「")
				b.WriteString(label)
				b.WriteString("」")
			}
			if k := strings.TrimSpace(p.PersonaKind); k != "" {
				b.WriteString(" persona_kind=")
				b.WriteString(k)
			}
			b.WriteString("\n")
		}
		return strings.TrimSpace(b.String())
	}
	if id := strings.TrimSpace(req.SpacePersonaID); id != "" {
		return "(参加者リスト未送信。判明している space_persona_id は会話相手のみ: " + id + ")"
	}
	return "（参加者の space_persona_id リストなし）"
}

func userFallbackFromParticipants(req ChatRequest) (id, name string) {
	for _, p := range req.Participants {
		if strings.EqualFold(strings.TrimSpace(p.PersonaKind), "self") {
			return strings.TrimSpace(p.SpacePersonaID), strings.TrimSpace(p.Label)
		}
	}
	return "", ""
}

func participantDisplayLabelMap(req ChatRequest) map[string]string {
	m := make(map[string]string)
	for _, p := range req.Participants {
		id := strings.TrimSpace(p.SpacePersonaID)
		if id == "" {
			continue
		}
		m[strings.ToLower(id)] = strings.TrimSpace(p.Label)
	}
	if sid := strings.ToLower(strings.TrimSpace(req.SpacePersonaID)); sid != "" {
		if strings.TrimSpace(m[sid]) == "" {
			m[sid] = speakerDisplayNameForRequest(req)
		}
	}
	return m
}

func (c *ChatHandler) ChatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if os.Getenv("FORCE_CHAT_API_ERROR") == "1" {
		http.Error(w, "forced error for testing", http.StatusServiceUnavailable)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if os.Getenv("OPENAI_API_KEY") == "" {
		http.Error(w, "Missing OpenAI API key", http.StatusInternalServerError)
		return
	}

	chatCtx := prompts.ChatContext{}
	applyPersonaFromRequest(chatCtx, req.Persona, req)
	chatCtx["tts_reading_styles"] = chatTTSReadingStylesPromptBlock(nil)
	chatCtx["dialogue_personas_digest"] = buildDialoguePersonasDigest(req)
	chatCtx["dialogue_participants_roster"] = formatDialogueParticipantsRoster(req)

	userFID, userFName := userFallbackFromParticipants(req)
	openAIMessages := chat.ClientMessagesToOpenAI(req.Messages, userFID, userFName)

	systemPrompt, err := prompts.LoadChatSystemMerged(chatCtx)
	if err != nil {
		logger.Error("Failed to load system prompt: %v", err)
		http.Error(w, "Failed to load system prompt", http.StatusInternalServerError)
		return
	}
	systemPrompt = chat.CleanSystemPrompt(systemPrompt)
	if devNote := config.ChatDevAssistSystemSuffix(); devNote != "" {
		systemPrompt = systemPrompt + " " + devNote
	}

	result, usage, err := chat.Chat(openAIMessages, systemPrompt, strings.TrimSpace(req.SpacePersonaID), speakerDisplayNameForRequest(req), participantDisplayLabelMap(req))
	if err != nil {
		logger.Error("Chat failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	turnID := strings.TrimSpace(r.Header.Get("X-Turn-Id"))
	openaiusage.LogForOpenAIHandler("chat", r.URL.Path, turnID, usage, openaiusage.UsagePresent(usage), systemPrompt, "")

	responseData := ChatResponse{
		Histories: result.Histories,
		Content:   result.Content,
		Script:    result.Script,
	}

	response, err := json.Marshal(responseData)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(response)
}
