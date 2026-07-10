package service

import (
	"encoding/json"
	"strconv"
	"strings"

	"voice-chat-api-go/logger"
)

type aivisStyleRow struct {
	Name            string `json:"name"`
	LocalID         int    `json:"local_id"`
	EngineSpeakerID *int   `json:"engine_speaker_id,omitempty"`
}

// aivisMultiPayload はカタログ同期のスピーカー単位 provider_payload（styles 配列あり）。
type aivisMultiPayload struct {
	AivmModelUUID   string `json:"aivm_model_uuid"`
	AivmSpeakerUUID string `json:"aivm_speaker_uuid"`
	SpeakerLocalID  int    `json:"speaker_local_id"`
	ModelName       string `json:"model_name,omitempty"`
	SpeakerName     string `json:"speaker_name,omitempty"`
	Category        string `json:"category,omitempty"`
	VoiceTimbre     string `json:"voice_timbre,omitempty"`
	Styles          []aivisStyleRow `json:"styles"`
}

// aivisApplyPayloadForSegment は voice_masters.provider_payload を解釈し、1 セグメント分の model / speaker を svc に載せる。
// styleLocalID はチャット AI が出す文字列（styles[].local_id と一致）。空なら local_id 最小のスタイル。
// styles が無いレガシー JSON は従来の ApplyProviderPayload にフォールバック。
func aivisApplyPayloadForSegment(svc *AivisService, raw json.RawMessage, styleLocalID string) {
	if len(raw) == 0 {
		return
	}
	var p aivisMultiPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		logger.Warn("aivis: provider_payload: %v", err)
		return
	}
	if len(p.Styles) == 0 {
		svc.ApplyProviderPayload(raw)
		return
	}
	if m := strings.TrimSpace(p.AivmModelUUID); m != "" {
		svc.modelUUID = m
	}
	st := pickAivisStyleEntry(p.Styles, styleLocalID)
	if st == nil {
		return
	}
	if st.EngineSpeakerID != nil {
		svc.speakerID = *st.EngineSpeakerID
	} else {
		svc.speakerID = st.LocalID
	}
}

func pickAivisStyleEntry(styles []aivisStyleRow, styleLocalID string) *aivisStyleRow {
	if len(styles) == 0 {
		return nil
	}
	want := strings.TrimSpace(styleLocalID)
	if want != "" {
		for i := range styles {
			if strconv.Itoa(styles[i].LocalID) == want {
				return &styles[i]
			}
		}
	}
	best := 0
	for i := 1; i < len(styles); i++ {
		if styles[i].LocalID < styles[best].LocalID {
			best = i
		}
	}
	return &styles[best]
}
