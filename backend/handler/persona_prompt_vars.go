package handler

import (
	"fmt"
	"strings"

	"voice-chat-api-go/prompts"
)

func applyPersonaFromRequest(chatCtx prompts.ChatContext, persona *PersonaPayload, req ChatRequest) {
	if persona != nil {
		name := strings.TrimSpace(persona.Name)
		if name == "" {
			name = participantLabelForSpeaker(req)
		}
		if name == "" {
			name = "アシスタント"
		}
		chatCtx["persona_name"] = name
		chatCtx["persona_personality"] = formatPersonalityText(persona.Personality)
		if s := strings.TrimSpace(persona.ResponseStyle); s != "" {
			chatCtx["response_style"] = s
		} else {
			chatCtx["response_style"] = "（未設定）"
		}
		if t := strings.TrimSpace(persona.Tempo); t != "" {
			chatCtx["tempo"] = t
		} else {
			chatCtx["tempo"] = "（未設定）"
		}
		return
	}

	chatCtx["persona_name"] = participantLabelForSpeaker(req)
	if chatCtx["persona_name"] == "" {
		chatCtx["persona_name"] = "アシスタント"
	}
	chatCtx["persona_personality"] = "（性格は未設定。会話から自然に振る舞ってください。）"
	chatCtx["response_style"] = "（未設定）"
	chatCtx["tempo"] = "（未設定）"
}

func participantLabelForSpeaker(req ChatRequest) string {
	speakerID := strings.TrimSpace(req.SpacePersonaID)
	for _, p := range req.Participants {
		if strings.TrimSpace(p.SpacePersonaID) == speakerID {
			if label := strings.TrimSpace(p.Label); label != "" {
				return label
			}
		}
	}
	return ""
}

// speakerDisplayNameForRequest は assistant 履歴 JSON の name に使う表示名。
// participants に話者（owned）が載らないデモ構成でも persona.name を正とする。
func speakerDisplayNameForRequest(req ChatRequest) string {
	if req.Persona != nil {
		if name := strings.TrimSpace(req.Persona.Name); name != "" {
			return name
		}
	}
	if label := participantLabelForSpeaker(req); label != "" {
		return label
	}
	return "アシスタント"
}

func formatPersonalityText(personality string) string {
	t := strings.TrimSpace(personality)
	if t == "" {
		return "（性格は未設定。会話から自然に振る舞ってください。）"
	}
	return t
}

func buildDialoguePersonasDigest(req ChatRequest) string {
	speakerID := strings.TrimSpace(req.SpacePersonaID)
	if req.Persona == nil && speakerID == "" {
		return "（AI ペルソナの設定がありません。参加者リスト・会話履歴を手がかりにしてください。）"
	}

	var b strings.Builder
	if speakerID != "" {
		b.WriteString("【space_persona_id: ")
		b.WriteString(speakerID)
		b.WriteString("】\n")
	}
	if label := participantLabelForSpeaker(req); label != "" {
		fmt.Fprintf(&b, "参加リスト上の表示名: %s\n", label)
	}
	if req.Persona != nil {
		name := strings.TrimSpace(req.Persona.Name)
		if name != "" {
			fmt.Fprintf(&b, "名前: %s\n", name)
		} else {
			b.WriteString("名前: アシスタント\n")
		}
		b.WriteString("性格・人物像:\n")
		b.WriteString(formatPersonalityText(req.Persona.Personality))
		if s := strings.TrimSpace(req.Persona.ResponseStyle); s != "" {
			fmt.Fprintf(&b, "\n応答スタイル: %s", s)
		}
	}
	return strings.TrimSpace(b.String())
}
