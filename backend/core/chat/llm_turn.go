package chat

// LLMUserTurn はバックエンドが OpenAI 呼び出し前に組み立てる直近ユーザ発話（内部用）。
type LLMUserTurn struct {
	Content string
	Lines   []AvatarScriptLine
}
