package hub

import (
	"sync"

	"voice-chat-api-go/logger"
)

// SSEEventFormat Server-Sent Events の 1 イベント書式
const SSEEventFormat = "event: %s\ndata: %s\n\n"

// Hub セッション単位の SSE 購読者と停止シグナルを管理する
type Hub struct {
	clients    map[string][]chan SynthesizeAudioResponse
	clientsMux sync.RWMutex
	stops      map[string]chan struct{}
}

var defaultHub = &Hub{
	clients: make(map[string][]chan SynthesizeAudioResponse),
	stops:   make(map[string]chan struct{}),
}

// Default プロセス内で共有する Hub（従来のグローバル sseManager と同じ寿命）
func Default() *Hub {
	return defaultHub
}

// AddClient 新しい SSE 接続を追加
func (h *Hub) AddClient(sessionID string) (chan SynthesizeAudioResponse, chan struct{}) {
	h.clientsMux.Lock()
	defer h.clientsMux.Unlock()

	clientChan := make(chan SynthesizeAudioResponse, 10)
	stopChan := make(chan struct{})
	h.clients[sessionID] = append(h.clients[sessionID], clientChan)
	h.stops[sessionID] = stopChan

	logger.Info("SSE connection added: sessionId=%s, connectionCount=%d", sessionID, len(h.clients[sessionID]))
	return clientChan, stopChan
}

// RemoveClient SSE 接続を削除
func (h *Hub) RemoveClient(sessionID string, clientChan chan SynthesizeAudioResponse) {
	h.clientsMux.Lock()
	defer h.clientsMux.Unlock()

	if clients, exists := h.clients[sessionID]; exists {
		for i, client := range clients {
			if client == clientChan {
				h.clients[sessionID] = append(clients[:i], clients[i+1:]...)
				close(clientChan)
				logger.Info("SSE connection removed: sessionId=%s, remainingConnections=%d", sessionID, len(h.clients[sessionID]))
				break
			}
		}
	}
}

// SendToSession セッションに紐づく全 SSE クライアントへ送る
func (h *Hub) SendToSession(sessionID string, response SynthesizeAudioResponse) {
	h.clientsMux.RLock()
	defer h.clientsMux.RUnlock()
	if clients, exists := h.clients[sessionID]; exists {
		for _, ch := range clients {
			ch <- response
		}
	}
}

// StopSession 停止チャネルを閉じる
func (h *Hub) StopSession(sessionID string) {
	h.clientsMux.Lock()
	defer h.clientsMux.Unlock()
	if stopChan, exists := h.stops[sessionID]; exists {
		close(stopChan)
		delete(h.stops, sessionID)
	}
}
