package chat

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestSegmentReplyToAvatarScript_empty(t *testing.T) {
	s := SegmentReplyToAvatarScript("", "0")
	if len(s.Lines) != 0 || s.StyleLocalID != "0" {
		t.Fatalf("%#v", s)
	}
}

func TestSegmentReplyToAvatarScript_invalidStyle(t *testing.T) {
	s := SegmentReplyToAvatarScript("こんにちは。", "abc")
	if s.StyleLocalID != "0" {
		t.Fatalf("style: %q", s.StyleLocalID)
	}
}

func TestSegmentReply_short(t *testing.T) {
	s := SegmentReplyToAvatarScript("短い。", "3")
	if len(s.Lines) < 1 || s.StyleLocalID != "3" {
		t.Fatalf("%#v", s)
	}
}

func TestSegmentReply_multipleSentences(t *testing.T) {
	text := "これは一文です。これは二文目です。そして三番目。"
	s := SegmentReplyToAvatarScript(text, "1")
	if len(s.Lines) == 0 {
		t.Fatal("no lines")
	}
	joined := strings.Join(linesTexts(s), "")
	if !strings.Contains(joined, "一文") || !strings.Contains(joined, "二文") {
		t.Fatalf("lines=%v", linesTexts(s))
	}
	for _, line := range s.Lines {
		if utf8.RuneCountInString(line.Text) > segmentMaxLineRunes*3 { // 結合後にやや長くなり得る
			t.Fatalf("line too long: %d %q", utf8.RuneCountInString(line.Text), line.Text)
		}
	}
}

func TestSegmentReply_noDanglingNda(t *testing.T) {
	text := "AGIは特定のタスクを超えて、幅広い分野で人間のように思考や学習ができるAIのことを指すんだ。"
	s := SegmentReplyToAvatarScript(text, "0")
	for _, l := range s.Lines {
		if strings.TrimSpace(l.Text) == "んだ。" {
			t.Fatalf("isolated んだ。: %#v", linesTexts(s))
		}
	}
}

func TestSegmentReply_literalBackslashN(t *testing.T) {
	// JSON 上は \\n だがアンマーシャル後はバックスラッシュ+n の2文字になり、splitSentences が改行として扱えない問題の救済
	s := SegmentReplyToAvatarScript("まず一句。\\n次の文です。", "0")
	txts := linesTexts(s)
	joined := strings.Join(txts, "")
	if strings.Contains(joined, "\\n") {
		t.Fatalf("literal \\n should be normalized away: %v", txts)
	}
	if !strings.Contains(joined, "次の文") {
		t.Fatalf("second part missing: %v", txts)
	}
}

func TestSegmentReply_longHardSplit(t *testing.T) {
	var b strings.Builder
	for i := 0; i < 120; i++ {
		b.WriteString("あ")
	}
	s := SegmentReplyToAvatarScript(b.String(), "2")
	if len(s.Lines) < 2 {
		t.Fatalf("expected split, got %d lines", len(s.Lines))
	}
}

func linesTexts(s AvatarScript) []string {
	out := make([]string, 0, len(s.Lines))
	for _, l := range s.Lines {
		out = append(out, l.Text)
	}
	return out
}
