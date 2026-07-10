package synthesis

// SynthesizeRequest 音声合成APIリクエストの共通構造体
type SynthesizeRequest struct {
	Message        string `json:"message"`
	Narrator       string `json:"narrator"`
	Speed          int    `json:"speed"`
	Pitch          int    `json:"pitch"`
	StyleLocalID   string `json:"style_local_id,omitempty"`
	Expression     string `json:"expression,omitempty"`
	Motion         string `json:"motion,omitempty"`
}
