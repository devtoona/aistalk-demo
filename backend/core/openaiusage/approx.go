package openaiusage

import (
	"os"
	"strings"

	"github.com/pkoukk/tiktoken-go"
)

func ApproxTextTokens(model, text string) int {
	if strings.TrimSpace(text) == "" {
		return 0
	}
	m := strings.TrimSpace(model)
	if m == "" {
		m = "gpt-4o-mini"
	}
	enc, err := tiktoken.EncodingForModel(m)
	if err != nil {
		enc, err = tiktoken.EncodingForModel("gpt-4")
		if err != nil {
			return 0
		}
	}
	return len(enc.Encode(text, nil, nil))
}

func resolveEncodingModel(component string) string {
	if component == "motion" {
		if m := strings.TrimSpace(os.Getenv("MOTION_MODEL")); m != "" {
			return m
		}
	}
	if m := strings.TrimSpace(os.Getenv("CHAT_MODEL")); m != "" {
		return m
	}
	if component == "motion" {
		return "gpt-4.1-nano"
	}
	return "gpt-4o-mini"
}
