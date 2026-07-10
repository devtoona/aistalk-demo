package prompts

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"voice-chat-api-go/config"
)

// ChatContext チャット用プロンプトに埋め込む変数。
// キーはプレースホルダ名（{{key}} の key 部分）に対応。
type ChatContext map[string]string

// LoadChatSystemMerged モジュールを順に読み込み、プレースホルダを置換してマージしたシステムプロンプトを返す。
// ベースファイルは CHAT_SYSTEM_PROFILE（config.ChatSystemBaseFileName）で single_chat / multi_chat / maker を切り替え。
func LoadChatSystemMerged(ctx ChatContext) (string, error) {
	dir := os.Getenv("PROMPTS_DIR")
	if dir == "" {
		dir = "prompts/files"
	}

	baseName := config.ChatSystemBaseFileName()
	var parts []string
	for _, name := range []string{baseName} {
		path := filepath.Join(dir, name)
		b, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue // ファイルがなければスキップ（後方互換）
			}
			return "", fmt.Errorf("read %s: %w", path, err)
		}
		s := replacePlaceholders(string(b), ctx)
		parts = append(parts, s)
	}

	if len(parts) == 0 {
		return "", fmt.Errorf("no chat prompt modules found in %s", dir)
	}

	return strings.TrimSpace(strings.Join(parts, "\n")), nil
}

var placeholderRE = regexp.MustCompile(`\{\{(\w+)\}\}`)

// replacePlaceholders s 内の {{key}} を ctx[key] で置換する。未定義は空文字。
func replacePlaceholders(s string, ctx ChatContext) string {
	return placeholderRE.ReplaceAllStringFunc(s, func(match string) string {
		key := match[2 : len(match)-2] // {{key}} から key を抽出
		if v, ok := ctx[key]; ok {
			return v
		}
		return ""
	})
}
