package config

import (
	"os"
	"strconv"
	"strings"
)

const (
	ChatTrimMaxTokensEnv     = "CHAT_TRIM_MAX_TOKENS"
	defaultChatTrimMaxTokens = 3000

	ChatDevAssistSystemEnv = "CHAT_DEV_ASSIST_SYSTEM"
	ChatSystemProfileEnv   = "CHAT_SYSTEM_PROFILE"
)

const chatDevAssistSystemNoteJP = "【開発用】いまは開発段階のため、開発者があなたに積極的に話しかけることがあります。その意図を汲み、協力的に応答しデバッグや検証をアシストしてください。"

func ChatTrimMaxTokens() int {
	v := os.Getenv(ChatTrimMaxTokensEnv)
	if v == "" {
		return defaultChatTrimMaxTokens
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return defaultChatTrimMaxTokens
	}
	return n
}

func ChatDevAssistSystemSuffix() string {
	v := strings.TrimSpace(os.Getenv(ChatDevAssistSystemEnv))
	if v == "" {
		return ""
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on":
		return chatDevAssistSystemNoteJP
	default:
		return ""
	}
}

func ChatSystemProfile() string {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(ChatSystemProfileEnv))) {
	case "maker":
		return "maker"
	case "multi_chat", "chat":
		return "multi_chat"
	default:
		return "single_chat"
	}
}

func ChatSystemBaseFileName() string {
	switch ChatSystemProfile() {
	case "maker":
		return "maker_system_base.txt"
	case "multi_chat":
		return "multi_chat_system_base.txt"
	default:
		return "single_chat_system_base.txt"
	}
}
