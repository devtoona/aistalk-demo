package hub

import (
	"encoding/base64"
	"strings"
	"time"

	"voice-chat-api-go/core/synthesis"
)

const defaultPlayExpression = "Neutral"
const defaultPlayMotion = "casual"

// SynthesizeAudioResponse 音声チャンクを SSE で送るときのペイロード（後続で汎用イベント型に寄せ可能）
type SynthesizeAudioResponse struct {
	Type        string `json:"type"`
	AudioData   string `json:"audioData"`
	Index       int    `json:"index"`
	Text        string `json:"text"`
	Timestamp   int64  `json:"timestamp"`
	TotalChunks int    `json:"totalChunks"`
	IsLast      bool   `json:"isLast"`
	SessionID   string `json:"sessionId"`
	Status      string `json:"status"`
	Expression  string `json:"expression,omitempty"`
	Motion      string `json:"motion,omitempty"`
}

// BuildPlayEvent 1 チャンク分の play イベント。synthErr != nil または audio == nil なら error 扱い（AudioData 空）
func BuildPlayEvent(sessionID string, index, total int, segment synthesis.SynthesizeRequest, audio []byte, synthErr error) SynthesizeAudioResponse {
	expr := strings.TrimSpace(segment.Expression)
	if expr == "" {
		expr = defaultPlayExpression
	}
	mot := strings.TrimSpace(segment.Motion)
	if mot == "" {
		mot = defaultPlayMotion
	}
	resp := SynthesizeAudioResponse{
		Type:        "play",
		Index:       index,
		Text:        segment.Message,
		Timestamp:   time.Now().UnixMilli(),
		TotalChunks: total,
		IsLast:      index == total-1,
		SessionID:   sessionID,
		Expression:  expr,
		Motion:      mot,
	}
	if synthErr != nil || audio == nil {
		resp.Status = "error"
		resp.AudioData = ""
		return resp
	}
	resp.Status = "success"
	resp.AudioData = base64.StdEncoding.EncodeToString(audio)
	return resp
}

// BuildCompleteEvent 全チャンク送信後の complete イベント
func BuildCompleteEvent(sessionID string) SynthesizeAudioResponse {
	return SynthesizeAudioResponse{
		Type:      "complete",
		SessionID: sessionID,
		Timestamp: time.Now().UnixMilli(),
		Status:    "success",
	}
}
