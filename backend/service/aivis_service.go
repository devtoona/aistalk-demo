package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"voice-chat-api-go/core/synthesis"
	"voice-chat-api-go/logger"
)

const (
	aivisModeLocal = "local"
	aivisModeCloud = "cloud"

	aivisDefaultBaseURL        = "https://api.aivis-project.com"
	aivisDefaultSynthesizePath = "/v1/tts/synthesize"
	aivisDefaultEngineURL      = "http://127.0.0.1:10101"
	aivisDefaultLocalSpeakerID = 888753762
	aivisDefaultModelUUID      = "a59cb814-0083-4369-8542-f51a29e72af7"
)

// AivisService Aivis: ローカルエンジン（VOICEVOX 互換 API）またはクラウド TTS
type AivisService struct {
	mode string // aivisModeLocal | aivisModeCloud
	// cloud
	apiURL    string
	apiKey    string
	modelUUID string // Cloud /v1/tts/synthesize の model_uuid。Aivis API 用の生 UUID（voice_masters.model_uuid の aivis: 複合キーは使わない）
	// local
	engineURL string
	speakerID int // /audio_query・/synthesis の speaker（数値）
	client    *http.Client
}

type aivisSynthesizePayload struct {
	ModelUUID              string  `json:"model_uuid"` // 必ず Aivis の aivm_model_uuid。DB 複合キーを渡さないこと
	Text                   string  `json:"text"`
	UseSSML                bool    `json:"use_ssml"`
	UseVolumeNormalizer    bool    `json:"use_volume_normalizer"`
	OutputFormat           string  `json:"output_format"`
	LeadingSilenceSeconds  float64 `json:"leading_silence_seconds"`
	TrailingSilenceSeconds float64 `json:"trailing_silence_seconds"`
}

func parseAivisMode() string {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("AIVIS_MODE")))
	if v == aivisModeCloud {
		return aivisModeCloud
	}
	return aivisModeLocal
}

// NewAivisService AIVIS_MODE: 未設定・空・cloud 以外 → local。cloud 時は AIVIS_API_KEY 必須。
// model / engine / speaker の既定値は voice_masters.provider_payload で上書きする。
func NewAivisService() *AivisService {
	mode := parseAivisMode()
	key := strings.TrimSpace(os.Getenv("AIVIS_API_KEY"))

	svc := &AivisService{
		mode:      mode,
		apiURL:    aivisBuildSynthesizeURL(),
		apiKey:    key,
		modelUUID: aivisDefaultModelUUID,
		engineURL: strings.TrimRight(aivisDefaultEngineURL, "/"),
		speakerID: aivisDefaultLocalSpeakerID,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}

	if mode == aivisModeCloud {
		if key == "" {
			logger.Warn("AIVIS_API_KEY is not set; Aivis cloud synthesis will fail")
		}
		logger.Info("Aivis: mode=cloud apiURL=%s", svc.apiURL)
	} else {
		logger.Info("Aivis: mode=local engineURL=%s speaker=%d", svc.engineURL, svc.speakerID)
	}
	return svc
}

// ApplyProviderPayload は voice_masters.provider_payload（JSON）で合成先を上書きする。
// Cloud では aivm_model_uuid、ローカルでは speaker_id（無ければ style_local_id）を参照する。
func (s *AivisService) ApplyProviderPayload(raw json.RawMessage) {
	cloud, speaker, haveCloud, haveSpeaker := parseAivisVoicePayload(raw)
	if haveCloud {
		s.modelUUID = cloud
	}
	if haveSpeaker {
		s.speakerID = speaker
	}
}

func parseAivisVoicePayload(raw json.RawMessage) (cloudModel string, localSpeaker int, haveCloud, haveSpeaker bool) {
	if len(raw) == 0 {
		return "", 0, false, false
	}
	t := strings.TrimSpace(string(raw))
	if t == "" || t == "{}" || t == "null" {
		return "", 0, false, false
	}
	var p struct {
		AivmModelUUID string `json:"aivm_model_uuid"`
		SpeakerID     *int   `json:"speaker_id"`
		StyleLocalID  *int   `json:"style_local_id"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		logger.Warn("aivis: provider_payload json: %v", err)
		return "", 0, false, false
	}
	cloudModel = strings.TrimSpace(p.AivmModelUUID)
	haveCloud = cloudModel != ""
	if p.SpeakerID != nil {
		return cloudModel, *p.SpeakerID, haveCloud, true
	}
	if p.StyleLocalID != nil {
		return cloudModel, *p.StyleLocalID, haveCloud, true
	}
	return cloudModel, 0, haveCloud, false
}

// SynthesizeAudio 1 セグメント分。cloud は MP3、local はエンジンが返す WAV
func (s *AivisService) SynthesizeAudio(req synthesis.SynthesizeRequest) ([]byte, error) {
	if s.mode == aivisModeCloud {
		return s.synthesizeCloud(req)
	}
	return s.synthesizeLocal(req)
}

func (s *AivisService) synthesizeCloud(req synthesis.SynthesizeRequest) ([]byte, error) {
	if strings.TrimSpace(s.apiKey) == "" {
		return nil, fmt.Errorf("aivis: AIVIS_API_KEY is not set (AIVIS_MODE=cloud)")
	}
	text := strings.TrimSpace(req.Message)
	if text == "" {
		return nil, fmt.Errorf("aivis: empty message")
	}
	payload := aivisSynthesizePayload{
		ModelUUID:              s.modelUUID,
		Text:                   text,
		UseSSML:                true,
		UseVolumeNormalizer:    true,
		OutputFormat:           "mp3",
		LeadingSilenceSeconds:  0.0,
		TrailingSilenceSeconds: 0.1,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("aivis: marshal: %w", err)
	}
	httpReq, err := http.NewRequest(http.MethodPost, s.apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("aivis: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("aivis: request: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("aivis: read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("aivis: status %d: %s", resp.StatusCode, truncateForLog(raw, 512))
	}
	return raw, nil
}

func (s *AivisService) synthesizeLocal(req synthesis.SynthesizeRequest) ([]byte, error) {
	text := strings.TrimSpace(req.Message)
	if text == "" {
		return nil, fmt.Errorf("aivis: empty message")
	}
	q, err := s.fetchLocalAudioQueryJSON(text)
	if err != nil {
		return nil, err
	}
	return s.synthesizeLocalFromQueryJSON(q)
}

func (s *AivisService) fetchLocalAudioQueryJSON(text string) ([]byte, error) {
	v := url.Values{}
	v.Set("speaker", strconv.Itoa(s.speakerID))
	v.Set("text", text)
	audioQueryURL := s.engineURL + "/audio_query?" + v.Encode()
	ga, err := http.NewRequest(http.MethodPost, audioQueryURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("aivis audio_query: %w", err)
	}
	resp, err := s.client.Do(ga)
	if err != nil {
		return nil, fmt.Errorf("aivis audio_query request: %w", err)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("aivis audio_query read: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("aivis audio_query status %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

func (s *AivisService) synthesizeLocalFromQueryJSON(queryJSON []byte) ([]byte, error) {
	if len(queryJSON) == 0 {
		return nil, fmt.Errorf("aivis: empty audio query json")
	}
	synQ := url.Values{}
	synQ.Set("speaker", strconv.Itoa(s.speakerID))
	synURL := s.engineURL + "/synthesis?" + synQ.Encode()
	sp, err := http.NewRequest(http.MethodPost, synURL, bytes.NewReader(queryJSON))
	if err != nil {
		return nil, fmt.Errorf("aivis synthesis: %w", err)
	}
	sp.Header.Set("Content-Type", "application/json")
	resp2, err := s.client.Do(sp)
	if err != nil {
		return nil, fmt.Errorf("aivis synthesis request: %w", err)
	}
	defer resp2.Body.Close()
	wav, err := io.ReadAll(resp2.Body)
	if err != nil {
		return nil, fmt.Errorf("aivis synthesis read: %w", err)
	}
	if resp2.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("aivis synthesis status %d: %s", resp2.StatusCode, string(wav))
	}
	return wav, nil
}

func truncateForLog(b []byte, max int) string {
	str := string(b)
	if len(str) <= max {
		return str
	}
	return str[:max] + "…"
}

func aivisBuildSynthesizeURL() string {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("AIVIS_BASE_URL")), "/")
	if base == "" {
		base = aivisDefaultBaseURL
	}
	path := strings.TrimSpace(os.Getenv("AIVIS_SYNTHESIZE_PATH"))
	if path == "" {
		path = aivisDefaultSynthesizePath
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return base + path
}
