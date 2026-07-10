package handler

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
)

type aivisStylePromptRow struct {
	Name    string `json:"name"`
	LocalID int    `json:"local_id"`
}

type aivisPayloadPrompt struct {
	Styles []aivisStylePromptRow `json:"styles"`
}

func chatTTSReadingStylesPromptBlock(payload []byte) string {
	if len(payload) == 0 {
		return defaultReadingStylesBlock("Aivis の provider_payload が空です。")
	}
	var p aivisPayloadPrompt
	if err := json.Unmarshal(payload, &p); err != nil || len(p.Styles) == 0 {
		return defaultReadingStylesBlock("Aivis の styles[] が読み取れないか空です。")
	}
	rows := append([]aivisStylePromptRow(nil), p.Styles...)
	sort.Slice(rows, func(i, j int) bool { return rows[i].LocalID < rows[j].LocalID })

	var b strings.Builder
	b.WriteString("**この応答全体でルートの `style_local_id` をちょうど1つだけ**選び、**すべての `lines` で同じ値**にすること。\n")
	b.WriteString("利用できる ID は次の**のみ**（**この一覧にない数字は出さない**）。セリフ全体のトーンに最も合うものを1つ選ぶ。\n")
	for _, r := range rows {
		name := strings.TrimSpace(r.Name)
		if name == "" {
			name = "（名称なし）"
		}
		b.WriteString("- `\"")
		b.WriteString(strconv.Itoa(r.LocalID))
		b.WriteString("\"` … ")
		b.WriteString(name)
		b.WriteString("\n")
	}
	b.WriteString("迷ったら最小の local_id（多くは `\"0\"`）を選ぶ。\n")
	return b.String()
}

func defaultReadingStylesBlock(contextLine string) string {
	var b strings.Builder
	b.WriteString("**この応答全体でルートの `style_local_id` をちょうど1つだけ**出力し、**すべての `lines` で同じ読み上げトーン**にすること（行ごとに変えない）。\n")
	b.WriteString(contextLine)
	b.WriteString(" **非負整数の文字列**のみ。一覧にない数字は禁止。迷ったら Aivis は `\"0\"`。\n")
	return b.String()
}
