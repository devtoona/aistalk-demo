package service

import "voice-chat-api-go/core/synthesis"

// AudioSynthesizeSessionRequest /api/event/tts/aivis/synthesize リクエストボディ
type AudioSynthesizeSessionRequest struct {
	SessionID     string                        `json:"sessionId"`
	VoiceMasterID string                        `json:"voice_master_id,omitempty"` // DB の voice_masters.id（Aivis / VoicePeak で provider_payload を参照）
	StyleLocalID  string                        `json:"style_local_id,omitempty"`  // VoicePeak: provider_payload.style の local_id（空なら "5" ノーマル）
	Segments      []synthesis.SynthesizeRequest `json:"segments"`
}
