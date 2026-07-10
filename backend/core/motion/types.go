package motion

// LineOut はセグメント1つ分。Text は delegate_segmentation 時にモデルが分割した読み上げ文。従来モードでは HTTP ハンドラがリクエストの text で埋める。
type LineOut struct {
	Text       string `json:"text,omitempty"`
	Motion     string `json:"motion"`
	Expression string `json:"expression"`
}

// ScriptOut OpenAI structured output のルート（lines のみ）。
type ScriptOut struct {
	Lines []LineOut `json:"lines"`
}
