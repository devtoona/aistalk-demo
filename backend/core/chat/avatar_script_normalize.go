package chat

import "strings"

// FilterSpeakableAvatarScript は trim 後に本文が空の行を除き、各行の Text を TrimSpace する。
// POST /api/chat の script はこの結果を返す（TTS・モーション・表示で行数を揃える）。
func FilterSpeakableAvatarScript(s AvatarScript) AvatarScript {
	out := make([]AvatarScriptLine, 0, len(s.Lines))
	for _, line := range s.Lines {
		t := strings.TrimSpace(line.Text)
		if t == "" {
			continue
		}
		out = append(out, AvatarScriptLine{Text: t})
	}
	return AvatarScript{Lines: out, StyleLocalID: s.StyleLocalID}
}
