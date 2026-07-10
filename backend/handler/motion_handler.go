package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"unicode/utf8"

	"voice-chat-api-go/core/chat"
	"voice-chat-api-go/core/motion"
	"voice-chat-api-go/core/openaiusage"
	"voice-chat-api-go/logger"
	"voice-chat-api-go/prompts"
)

// MotionHandler POST /api/avatar/motion
type MotionHandler struct{}

func NewMotionHandler() *MotionHandler {
	return &MotionHandler{}
}

const motionLastUserMessageMaxRunes = 1200

type motionInferRequest struct {
	Lines []struct {
		Text string `json:"text"`
	} `json:"lines"`
	DelegateSegmentation bool   `json:"delegate_segmentation,omitempty"`
	StyleLocalID         string `json:"style_local_id,omitempty"`
	LastUserMessage      string `json:"last_user_message,omitempty"`
}

type motionInferResponse struct {
	Lines        []motion.LineOut `json:"lines"`
	UsedFallback bool             `json:"used_fallback"`
}

type motionUserPayload struct {
	Segments             []motionSegmentIn `json:"segments"`
	DelegateSegmentation bool              `json:"delegate_segmentation,omitempty"`
	StyleLocalID         string            `json:"style_local_id,omitempty"`
	LastUserMessage      string            `json:"last_user_message,omitempty"`
}

type motionSegmentIn struct {
	Index int    `json:"index"`
	Text  string `json:"text"`
}

func (h *MotionHandler) PostInfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if os.Getenv("OPENAI_API_KEY") == "" {
		http.Error(w, `{"error":"OPENAI_API_KEY is not set"}`, http.StatusInternalServerError)
		return
	}

	var req motionInferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	n := len(req.Lines)
	delegate := req.DelegateSegmentation
	if n == 0 || n > motion.MaxLinesPerRequest {
		http.Error(w, `{"error":"lines must have 1..10 non-empty items"}`, http.StatusBadRequest)
		return
	}
	if delegate && n != 1 {
		http.Error(w, `{"error":"delegate_segmentation requires exactly one lines item"}`, http.StatusBadRequest)
		return
	}

	for i := range req.Lines {
		t := strings.TrimSpace(req.Lines[i].Text)
		if t == "" {
			http.Error(w, `{"error":"each lines[].text must be non-empty"}`, http.StatusBadRequest)
			return
		}
		maxRunes := motion.MaxTextChars
		if delegate {
			maxRunes = motion.MaxDelegateAssistantRunes
		}
		if utf8.RuneCountInString(t) > maxRunes {
			http.Error(w, `{"error":"each lines[].text exceeds max length"}`, http.StatusBadRequest)
			return
		}
	}

	segments := make([]motionSegmentIn, n)
	for i := range req.Lines {
		segments[i] = motionSegmentIn{
			Index: i,
			Text:  strings.TrimSpace(req.Lines[i].Text),
		}
	}
	lastUser := strings.TrimSpace(req.LastUserMessage)
	if lastUser != "" {
		runes := []rune(lastUser)
		if len(runes) > motionLastUserMessageMaxRunes {
			lastUser = string(runes[:motionLastUserMessageMaxRunes])
		}
	}
	userPayload := motionUserPayload{
		Segments:             segments,
		DelegateSegmentation: delegate,
		StyleLocalID:         strings.TrimSpace(req.StyleLocalID),
		LastUserMessage:      lastUser,
	}
	userBytes, err := json.Marshal(userPayload)
	if err != nil {
		http.Error(w, `{"error":"encode user payload"}`, http.StatusInternalServerError)
		return
	}
	userContent := string(userBytes)

	systemRaw, err := prompts.LoadAvatarMotionSystem()
	if err != nil {
		logger.Error("LoadAvatarMotionSystem: %v", err)
		http.Error(w, `{"error":"motion system prompt unavailable"}`, http.StatusInternalServerError)
		return
	}
	systemPrompt := chat.CleanSystemPrompt(systemRaw)

	var resp motionInferResponse
	script, usage, err := motion.Infer(systemPrompt, userContent)
	if openaiusage.UsagePresent(usage) {
		turnID := strings.TrimSpace(r.Header.Get("X-Turn-Id"))
		openaiusage.LogForOpenAIHandler("motion", r.URL.Path, turnID, usage, true, systemPrompt, userContent)
	}
	ok := err == nil && script != nil
	if ok && !delegate {
		ok = len(script.Lines) == n
	}
	if ok && delegate {
		nOut := len(script.Lines)
		ok = nOut >= 1 && nOut <= motion.MaxLinesPerRequest
	}
	if !ok {
		if err != nil {
			logger.Warn("motion infer failed (using fallback): %v", err)
		} else if script == nil {
			logger.Warn("motion infer: nil script (using fallback)")
		} else if delegate {
			logger.Warn("motion infer: delegate invalid line count got=%d", len(script.Lines))
		} else {
			logger.Warn("motion infer: line count mismatch want=%d got=%d", n, len(script.Lines))
		}
		if delegate {
			full := strings.TrimSpace(req.Lines[0].Text)
			resp.Lines = motion.DefaultLinesWithTexts([]string{full})
		} else {
			resp.Lines = motion.DefaultLines(n)
		}
		resp.UsedFallback = true
	} else {
		norm := motion.NormalizeLines(script.Lines)
		if delegate {
			var kept []motion.LineOut
			for _, ln := range norm {
				ts := strings.TrimSpace(ln.Text)
				if ts == "" {
					continue
				}
				ln.Text = ts
				kept = append(kept, ln)
			}
			if len(kept) < 1 || len(kept) > motion.MaxLinesPerRequest {
				full := strings.TrimSpace(req.Lines[0].Text)
				resp.Lines = motion.DefaultLinesWithTexts([]string{full})
				resp.UsedFallback = true
			} else {
				resp.Lines = kept
				resp.UsedFallback = false
			}
		} else {
			for i := range norm {
				if i < len(req.Lines) {
					norm[i].Text = strings.TrimSpace(req.Lines[i].Text)
				}
			}
			resp.Lines = norm
			resp.UsedFallback = false
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Error("motion response encode: %v", err)
	}
}
