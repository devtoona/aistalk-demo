package chat

import (
	"encoding/json"
	"strings"
)

// ConversationSituationDetail クライアント履歴 JSON（ParseChatTurnContent）内の conversation_situation 用。
// サーバは user の role/content に会話状況を追記しない。レガシー互換のパースのみ。
type ConversationSituationDetail struct {
	TurnCandidates   []string `json:"turn_candidates"`
	SituationSummary string   `json:"situation_summary"`
}

// AssistantHistoryItem assistant の会話履歴 1 話者分（複数ブロックは配列で並べる）。
type AssistantHistoryItem struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Message string `json:"message"`
}

// ChatTurnContent 会話履歴の messages[].content に埋め込む JSON（OpenAI API は文字列のみのため JSON 文字列として渡す）。
// conversation_situation はクライアントが付けたレガシー用。正規化後のモデル入力からは除去する。
type ChatTurnContent struct {
	ID                    string                       `json:"id"`
	Name                  string                       `json:"name"`
	Message               string                       `json:"message"`
	ConversationSituation *ConversationSituationDetail `json:"conversation_situation,omitempty"`
}

// MarshalChatTurnJSON id は space_persona_id（不明時は空）、name は表示名、message は発話本文。
func MarshalChatTurnJSON(id, name, message string) (string, error) {
	c := ChatTurnContent{
		ID:      strings.TrimSpace(id),
		Name:    strings.TrimSpace(name),
		Message: strings.TrimSpace(message),
	}
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ParseChatTurnContent content が当形式の JSON なら true と中身を返す。
// conversation_situation がオブジェクトまたは旧来の文字列（要約のみ）の両方を解釈する。
func ParseChatTurnContent(raw string) (ChatTurnContent, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw[0] != '{' {
		return ChatTurnContent{}, false
	}
	var aux struct {
		ID    string          `json:"id"`
		Name  string          `json:"name"`
		Msg   string          `json:"message"`
		SitIn json.RawMessage `json:"conversation_situation,omitempty"`
	}
	if err := json.Unmarshal([]byte(raw), &aux); err != nil {
		return ChatTurnContent{}, false
	}
	c := ChatTurnContent{
		ID:      strings.TrimSpace(aux.ID),
		Name:    strings.TrimSpace(aux.Name),
		Message: strings.TrimSpace(aux.Msg),
	}
	if len(aux.SitIn) == 0 {
		return c, true
	}
	switch aux.SitIn[0] {
	case '"':
		var s string
		if err := json.Unmarshal(aux.SitIn, &s); err == nil && strings.TrimSpace(s) != "" {
			c.ConversationSituation = &ConversationSituationDetail{SituationSummary: strings.TrimSpace(s)}
		}
	case '{':
		var d ConversationSituationDetail
		if err := json.Unmarshal(aux.SitIn, &d); err == nil {
			c.ConversationSituation = &d
		}
	}
	return c, true
}

// NormalizeMessageContentForModel モデル入力用に content を正規化する。
// user … 発話本文のみの平文（クライアントが JSON を送った場合は message を抽出。conversation_situation は除去）。
// assistant … `{id,name,message}` の JSON 配列（旧単一オブジェクトは 1 要素配列に変換）。
// クライアントが user JSON に conversation_situation を含めていても、正規化で本文からは落とす。
func NormalizeMessageContentForModel(role, raw, userFallbackID, userFallbackName string) string {
	raw = strings.TrimSpace(raw)
	if role == "assistant" {
		return normalizeAssistantContentForModel(raw)
	}
	if c, ok := ParseChatTurnContent(raw); ok {
		return strings.TrimSpace(c.Message)
	}
	return raw
}

func normalizeAssistantContentForModel(raw string) string {
	if raw == "" {
		b, _ := json.Marshal([]AssistantHistoryItem{{}})
		return string(b)
	}
	switch raw[0] {
	case '[':
		var items []AssistantHistoryItem
		if err := json.Unmarshal([]byte(raw), &items); err == nil {
			if len(items) == 0 {
				items = []AssistantHistoryItem{{}}
			}
			out := make([]AssistantHistoryItem, 0, len(items))
			for _, it := range items {
				out = append(out, AssistantHistoryItem{
					ID:      strings.TrimSpace(it.ID),
					Name:    strings.TrimSpace(it.Name),
					Message: strings.TrimSpace(it.Message),
				})
			}
			b, _ := json.Marshal(out)
			return string(b)
		}
	case '{':
		if c, ok := ParseChatTurnContent(raw); ok {
			b, _ := json.Marshal([]AssistantHistoryItem{{
				ID:      strings.TrimSpace(c.ID),
				Name:    strings.TrimSpace(c.Name),
				Message: strings.TrimSpace(c.Message),
			}})
			return string(b)
		}
	}
	b, _ := json.Marshal([]AssistantHistoryItem{{Message: raw}})
	return string(b)
}
