package service

import "voice-chat-api-go/hub"

// SessionPublisher 合成結果を SSE で配信する抽象（*hub.Hub が満たす）
type SessionPublisher interface {
	SendToSession(sessionID string, ev hub.SynthesizeAudioResponse)
}
