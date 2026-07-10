package chat

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// サーバ側の TTS 用分割（チャット LLM は reply のみ返す想定）。
const (
	segmentMaxLineRunes       = 44
	segmentSmartSplitMinRunes = 30 // 長文時、自然区切りを探す先頭ブロックの最小ルーン数
	segmentMaxLines           = 16
	// mergeIfTailRunesLTE 以下の行、または語尾パターンに一致する行は前行へ結合する
	mergeIfTailRunesLTE = 6
)

var reDigitsOnly = regexp.MustCompile(`^[0-9]+$`)

// SegmentReplyToAvatarScript はプレーンな reply を AvatarScript.lines に分割する。
func SegmentReplyToAvatarScript(reply, styleLocalID string) AvatarScript {
	sid := strings.TrimSpace(styleLocalID)
	if !reDigitsOnly.MatchString(sid) {
		sid = "0"
	}
	lines := segmentReplyLines(strings.TrimSpace(reply))
	if len(lines) == 0 {
		return AvatarScript{Lines: nil, StyleLocalID: sid}
	}
	return AvatarScript{Lines: lines, StyleLocalID: sid}
}

func segmentReplyLines(s string) []AvatarScriptLine {
	if s == "" {
		return nil
	}
	s = normalizeReplyLiteralEscapes(s)
	segs := splitSentences(s)
	packed := packSentences(segs, segmentMaxLineRunes)
	var out []AvatarScriptLine
	for _, p := range packed {
		for _, part := range splitLongChunk(p, segmentMaxLineRunes) {
			t := strings.TrimSpace(part)
			if t == "" {
				continue
			}
			out = append(out, AvatarScriptLine{Text: t})
		}
	}
	out = mergeDanglingTailLines(out)
	out = collapseToMaxLines(out, segmentMaxLines)
	return out
}

// normalizeReplyLiteralEscapes は LLM が JSON の reply に「実改行」ではなくリテラル \n / \r / \t を
// 書いた場合の救済。区切り・分割の前に適用する。
func normalizeReplyLiteralEscapes(s string) string {
	if s == "" {
		return s
	}
	s = strings.ReplaceAll(s, "\\r\\n", "\r\n")
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\r", "\r")
	s = strings.ReplaceAll(s, "\\t", "\t")
	return s
}

// splitSentences は句点・感嘆・疑問・改行を主な境界に分割（境界文字は前の要素に含める）。
func splitSentences(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	var parts []string
	var b strings.Builder
	flush := func() {
		t := strings.TrimSpace(b.String())
		b.Reset()
		if t != "" {
			parts = append(parts, t)
		}
	}
	for _, r := range s {
		if r == '\n' {
			flush()
			continue
		}
		b.WriteRune(r)
		if r == '。' || r == '！' || r == '？' || r == '…' {
			flush()
		}
	}
	flush()
	if len(parts) == 0 {
		return []string{strings.TrimSpace(s)}
	}
	return parts
}

func packSentences(sentences []string, maxRunes int) []string {
	if maxRunes < 1 {
		return sentences
	}
	var out []string
	var cur strings.Builder
	curRunes := 0
	flushCur := func() {
		t := strings.TrimSpace(cur.String())
		cur.Reset()
		curRunes = 0
		if t != "" {
			out = append(out, t)
		}
	}
	for _, sent := range sentences {
		sent = strings.TrimSpace(sent)
		if sent == "" {
			continue
		}
		n := utf8.RuneCountInString(sent)
		if curRunes == 0 {
			if n <= maxRunes {
				cur.WriteString(sent)
				curRunes = n
				continue
			}
			// 単文が長い → そのまま後段で hard split
			flushCur()
			out = append(out, sent)
			continue
		}
		if curRunes+n <= maxRunes {
			cur.WriteString(sent)
			curRunes += n
			continue
		}
		flushCur()
		if n <= maxRunes {
			cur.WriteString(sent)
			curRunes = n
		} else {
			out = append(out, sent)
		}
	}
	flushCur()
	return out
}

// splitLongChunk は maxRunes 超の塊だけ再帰分割する。maxRunes 以下はそのまま 1 要素。
func splitLongChunk(s string, maxRunes int) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	rs := []rune(s)
	if maxRunes < 1 {
		return []string{s}
	}
	if len(rs) <= maxRunes {
		return []string{s}
	}
	return splitLongRunesRec(rs, maxRunes)
}

func splitLongRunesRec(rs []rune, max int) []string {
	if len(rs) <= max {
		t := strings.TrimSpace(string(rs))
		if t == "" {
			return nil
		}
		return []string{t}
	}
	minW := segmentSmartSplitMinRunes
	if minW < 1 {
		minW = 1
	}
	if minW > max {
		minW = max
	}
	if cut := findPreferredCutInWindow(rs, minW, max); cut > 0 {
		first := strings.TrimSpace(string(rs[:cut]))
		rest := trimRunesLeft(rs[cut:])
		var out []string
		if first != "" {
			out = append(out, first)
		}
		if len(rest) > 0 {
			out = append(out, splitLongRunesRec(rest, max)...)
		}
		return out
	}
	// フォールバック: 従来の 44 相当位置＋直前最大10ルーンで読点・空白のみ
	end := max
	if end > len(rs) {
		end = len(rs)
	} else {
		scanStart := end - 1
		scanLow := end - 10
		if scanLow < 0 {
			scanLow = 0
		}
		for j := scanStart; j >= scanLow; j-- {
			switch rs[j] {
			case '、', '，', ',', ' ':
				end = j + 1
				goto foundFallback
			}
		}
	foundFallback:
	}
	first := strings.TrimSpace(string(rs[:end]))
	rest := trimRunesLeft(rs[end:])
	var out2 []string
	if first != "" {
		out2 = append(out2, first)
	}
	if len(rest) > 0 {
		out2 = append(out2, splitLongRunesRec(rest, max)...)
	}
	return out2
}

func trimRunesLeft(rs []rune) []rune {
	i := 0
	for i < len(rs) && (rs[i] == ' ' || rs[i] == '\t' || rs[i] == '\n' || rs[i] == '\r') {
		i++
	}
	return rs[i:]
}

// findPreferredCutInWindow は先頭ブロック長が [minLen, maxLen] の範囲で、優先度に従い最も後ろの区切りを返す（exclusive cut index）。見つからなければ 0。
func findPreferredCutInWindow(rs []rune, minLen, maxLen int) int {
	if len(rs) <= maxLen {
		return 0
	}
	start := minLen
	if start < 1 {
		start = 1
	}
	end := maxLen
	if end > len(rs) {
		end = len(rs)
	}
	// 1) 句点
	for cut := end; cut >= start; cut-- {
		switch rs[cut-1] {
		case '。', '！', '？', '…':
			return cut
		}
	}
	// 2) 読点（区切りは読点を前ブロックに含める）
	for cut := end; cut >= start; cut-- {
		switch rs[cut-1] {
		case '、', '，', ',':
			return cut
		}
	}
	// 3) 助詞で終わる位置（助詞を前ブロックに含める）
	for cut := end; cut >= start; cut-- {
		switch rs[cut-1] {
		case 'は', 'が', 'を', 'に', 'で', 'と', 'の', 'も', 'へ':
			return cut
		}
	}
	return 0
}

var danglingTailPatterns = []string{
	"んだ。", "です。", "ます。", "ない。", "でした。", "ません。", "だよ。", "だね。", "かな。", "ぞ。", "わ。", "よ。", "ね。",
}

func mergeDanglingTailLines(lines []AvatarScriptLine) []AvatarScriptLine {
	if len(lines) < 2 {
		return lines
	}
	out := []AvatarScriptLine{lines[0]}
	for i := 1; i < len(lines); i++ {
		cur := strings.TrimSpace(lines[i].Text)
		if cur == "" {
			continue
		}
		if shouldMergeTailToPrevious(cur) {
			prev := len(out) - 1
			if prev < 0 {
				out = append(out, AvatarScriptLine{Text: cur})
				continue
			}
			out[prev] = AvatarScriptLine{Text: strings.TrimSpace(out[prev].Text + cur)}
			continue
		}
		out = append(out, AvatarScriptLine{Text: cur})
	}
	return out
}

func shouldMergeTailToPrevious(t string) bool {
	n := utf8.RuneCountInString(t)
	if n <= mergeIfTailRunesLTE {
		return true
	}
	for _, p := range danglingTailPatterns {
		if t == p {
			return true
		}
	}
	return false
}

func collapseToMaxLines(lines []AvatarScriptLine, max int) []AvatarScriptLine {
	if max < 1 || len(lines) <= max {
		return lines
	}
	out := append([]AvatarScriptLine(nil), lines...)
	for len(out) > max {
		bestI := 0
		bestCost := int(^uint(0) >> 1)
		for i := 0; i < len(out)-1; i++ {
			c := utf8.RuneCountInString(out[i].Text) + utf8.RuneCountInString(out[i+1].Text)
			if c < bestCost {
				bestCost = c
				bestI = i
			}
		}
		merged := strings.TrimSpace(out[bestI].Text + out[bestI+1].Text)
		out[bestI] = AvatarScriptLine{Text: merged}
		out = append(out[:bestI+1], out[bestI+2:]...)
	}
	return out
}
