package motion

import "strings"

const (
	defaultMotion     = "casual"
	defaultExpression = "Neutral"
)

var allowedExpressions = map[string]struct{}{
	"Neutral": {}, "Happiness": {}, "Sad": {}, "Angry": {}, "Surprised": {},
	"Joy": {}, "None": {},
	"Tsundere": {}, "Hanikami": {}, "KomariDere": {}, "Saiaku": {}, "Nemui": {},
	"Koakuma": {}, "Jitome": {},
	"Amae": {},
}

// DefaultLine returns casual + Neutral。
func DefaultLine() LineOut {
	return LineOut{
		Motion:     defaultMotion,
		Expression: defaultExpression,
	}
}

// DefaultLines returns n 件のデフォルト（モーション API 失敗時）。
func DefaultLines(n int) []LineOut {
	if n <= 0 {
		return nil
	}
	out := make([]LineOut, n)
	d := DefaultLine()
	for i := range out {
		out[i] = d
	}
	return out
}

// DefaultLinesWithTexts は delegate 失敗時など、読み上げ文付きの casual + Neutral を返す。
func DefaultLinesWithTexts(texts []string) []LineOut {
	if len(texts) == 0 {
		return nil
	}
	d := DefaultLine()
	out := make([]LineOut, len(texts))
	for i, t := range texts {
		out[i] = LineOut{Text: strings.TrimSpace(t), Motion: d.Motion, Expression: d.Expression}
	}
	return out
}

// NormalizeLine は enum 外をデフォルトに置き換える。Text はトリムのみ（空はそのまま）。
func NormalizeLine(in LineOut) LineOut {
	m := canonicalMotion(strings.TrimSpace(in.Motion))
	if _, ok := allowedSemanticMotionSet[m]; !ok {
		m = defaultMotion
	}
	ex := strings.TrimSpace(in.Expression)
	// 旧応答・手入力の Amae2 は API 上は Amae のみ（再生バリエーションはクライアント）
	if strings.EqualFold(ex, "Amae2") {
		ex = "Amae"
	}
	if _, ok := allowedExpressions[ex]; !ok {
		ex = defaultExpression
	}
	return LineOut{
		Text:       strings.TrimSpace(in.Text),
		Motion:     m,
		Expression: ex,
	}
}

// NormalizeLines applies NormalizeLine to each element.
func NormalizeLines(lines []LineOut) []LineOut {
	out := make([]LineOut, len(lines))
	for i := range lines {
		out[i] = NormalizeLine(lines[i])
	}
	return out
}
