package service

import (
	"os"
	"strings"

	"voice-chat-api-go/hub"
	"voice-chat-api-go/logger"
)

// RunAivis Aivis によるセグメント連続合成と SSE 通知（DB なし）
func RunAivis(pub SessionPublisher, req AudioSynthesizeSessionRequest) {
	modelUUID := strings.TrimSpace(os.Getenv("AIVIS_MODEL_UUID"))
	if vm := strings.TrimSpace(req.VoiceMasterID); vm != "" && modelUUID == "" {
		modelUUID = vm
	}

	n := len(req.Segments)
	for index, item := range req.Segments {
		styleID := strings.TrimSpace(item.StyleLocalID)
		if styleID == "" {
			styleID = strings.TrimSpace(req.StyleLocalID)
		}
		svc := NewAivisService()
		if modelUUID != "" {
			svc.modelUUID = modelUUID
		}
		audioData, err := svc.SynthesizeAudio(item)
		resp := hub.BuildPlayEvent(req.SessionID, index, n, item, audioData, err)
		if err != nil {
			logger.Error("Aivis synthesis error (chunk %d/%d): %v", index+1, n, err)
		} else {
			logger.Info("Aivis synthesis completed (chunk %d/%d): audioSize=%d", index+1, n, len(audioData))
		}
		pub.SendToSession(req.SessionID, resp)
	}
	pub.SendToSession(req.SessionID, hub.BuildCompleteEvent(req.SessionID))
	logger.Info("Aivis synthesis processing completed for session: %s", req.SessionID)
}
