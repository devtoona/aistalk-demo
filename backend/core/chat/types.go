package chat

import (
	"time"
)

// Message OpenAI API および履歴レスポンス用の最小形（role + content のみをモデルに渡す）。
// OpenAI 向け: user は平文（サーバは会話状況を本文に追記しない）。assistant は JSON 配列など正規化後の文字列。
type Message struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

// ChatClientMessage POST /api/chat の messages[] の各要素。
// content は {"id","name","message"} の JSON 文字列を推奨。conversation_situation が付いていてもモデル入力では除去。プレーン文の user はサーバが participants の self でラップする。lines は任意。
type ChatClientMessage struct {
	Role      string             `json:"role"`
	Content   string             `json:"content"`
	CreatedAt time.Time          `json:"createdAt"`
	Lines     []AvatarScriptLine `json:"lines,omitempty"`
}

// Result Chat の実行結果
type Result struct {
	Histories []Message
	Content   string
	// HTTP レスポンスの script（lines は trim 済み・空行なし。話者はリクエストの space_persona_id）。TTS 用はクライアントが script から GenerateVoice 用配列に写すのみ。
	Script *AvatarScript
}

// AvatarScript GPT の構造化出力と POST /api/chat の script が同型（lines + style_local_id）。
type AvatarScript struct {
	Lines []AvatarScriptLine `json:"lines"`
	// 読み上げスタイル（非負整数の文字列）。この応答のすべての lines で共通。VoicePeak / Aivis とも provider_payload 内の local_id と一致させる。
	StyleLocalID string `json:"style_local_id"`
}

// AvatarScriptLine 1 セグメント分のテキスト（聞き手は常にプレイヤー想定）。読み上げスタイルはルート style_local_id。モーション・表情は POST /api/avatar/motion。
type AvatarScriptLine struct {
	Text string `json:"text"`
}
