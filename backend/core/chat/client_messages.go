package chat

// ClientMessagesToOpenAI クライアントメッセージを OpenAI messages 用に変換する（lines はモデル入力に含めない）。
// user … 平文（JSON なら message を抽出）。assistant … JSON 配列等に正規化。
func ClientMessagesToOpenAI(in []ChatClientMessage, userFallbackID, userFallbackName string) []Message {
	if len(in) == 0 {
		return nil
	}
	out := make([]Message, len(in))
	for i := range in {
		out[i] = Message{
			Role:      in[i].Role,
			Content:   NormalizeMessageContentForModel(in[i].Role, in[i].Content, userFallbackID, userFallbackName),
			CreatedAt: in[i].CreatedAt,
		}
	}
	return out
}
