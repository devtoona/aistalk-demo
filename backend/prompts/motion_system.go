package prompts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const avatarMotionSystemFile = "avatar_motion_system_base.txt"

// LoadAvatarMotionSystem モーション推論専用のシステムプロンプト（1 ファイル）。
func LoadAvatarMotionSystem() (string, error) {
	dir := os.Getenv("PROMPTS_DIR")
	if dir == "" {
		dir = "prompts/files"
	}
	path := filepath.Join(dir, avatarMotionSystemFile)
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("motion system prompt not found: %s", path)
		}
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	return strings.TrimSpace(string(b)), nil
}
