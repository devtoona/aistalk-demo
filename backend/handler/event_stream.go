package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"voice-chat-api-go/hub"
	"voice-chat-api-go/logger"
	"voice-chat-api-go/service"
)

// EventStreamHandler イベントストリーム（SSE）と TTS 受付の HTTP 層
type EventStreamHandler struct {
	h *hub.Hub
}

// NewEventStreamHandler 共有 Hub を使う
func NewEventStreamHandler() *EventStreamHandler {
	return &EventStreamHandler{h: hub.Default()}
}

func parseAudioSynthesizeRequest(r *http.Request) (service.AudioSynthesizeSessionRequest, string) {
	var req service.AudioSynthesizeSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, "Invalid request body"
	}
	if req.Segments == nil {
		return req, "segments is required"
	}
	if strings.TrimSpace(req.SessionID) == "" {
		return req, "sessionId is required"
	}
	return req, ""
}

func writeTTSRequestAcceptedJSON(w http.ResponseWriter, sessionID string) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "accepted",
		"message":   "音声合成リクエストを受信しました",
		"sessionId": sessionID,
		"timestamp": time.Now().UnixMilli(),
	})
}

// SynthesisSessionStartHandler SSE セッション開始
func (c *EventStreamHandler) SynthesisSessionStartHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		http.Error(w, "sessionId is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	clientChan, stopChan := c.h.AddClient(sessionID)
	defer c.h.RemoveClient(sessionID, clientChan)

	logger.Info("SSE connection started: sessionId=%s", sessionID)

	msg := map[string]interface{}{
		"type":      "connected",
		"sessionId": sessionID,
		"timestamp": time.Now().UnixMilli(),
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		logger.Error("SSE JSON encode error: %v", err)
		return
	}

	fmt.Fprintf(w, hub.SSEEventFormat, "connected", string(jsonData))
	w.(http.Flusher).Flush()

	logger.Debug("SSE connection confirmation sent for sessionId: %s", sessionID)

	for {
		select {
		case response := <-clientChan:
			jsonData, err := json.Marshal(response)
			if err != nil {
				logger.Error("SSE JSON encode error for response: %v", err)
				continue
			}

			eventName := response.Type
			if eventName == "" {
				eventName = "play"
			}
			fmt.Fprintf(w, hub.SSEEventFormat, eventName, string(jsonData))
			w.(http.Flusher).Flush()

		case <-stopChan:
			logger.Info("SSE connection ended (stop signal): sessionId=%s", sessionID)
			return

		case <-r.Context().Done():
			logger.Info("SSE connection ended (context done): sessionId=%s", sessionID)
			return
		}
	}
}

// SynthesisSessionStopHandler セッション終了（クライアント離脱など）
func (c *EventStreamHandler) SynthesisSessionStopHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID string `json:"sessionId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode synthesis session stop request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		logger.Warn("Synthesis session stop request missing sessionId")
		http.Error(w, "sessionId is required", http.StatusBadRequest)
		return
	}

	c.h.StopSession(req.SessionID)

	logger.Info("Audio synthesis session stopped: sessionId=%s", req.SessionID)

	response := map[string]interface{}{
		"type":      "synthesis_session_stopped",
		"sessionId": req.SessionID,
		"timestamp": time.Now().UnixMilli(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SynthesizeAivisAudioHandler Aivis 合成（非同期受付）
func (c *EventStreamHandler) SynthesizeAivisAudioHandler(w http.ResponseWriter, r *http.Request) {
	req, errMsg := parseAudioSynthesizeRequest(r)
	if errMsg != "" {
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}
	writeTTSRequestAcceptedJSON(w, req.SessionID)
	go service.RunAivis(c.h, req)
}
