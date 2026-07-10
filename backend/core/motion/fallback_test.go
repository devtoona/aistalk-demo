package motion

import "testing"

func TestNormalizeLine_legacyAnimatorPathToSemanticCasual(t *testing.T) {
	got := NormalizeLine(LineOut{Motion: "KA_Idle.React_SM.Laugh", Expression: "Neutral"})
	if got.Motion != "casual" {
		t.Fatalf("motion: %q", got.Motion)
	}
}

func TestNormalizeLine_preservesText(t *testing.T) {
	got := NormalizeLine(LineOut{Text: " こんにちは ", Motion: "casual", Expression: "Neutral"})
	if got.Text != "こんにちは" || got.Motion != "casual" {
		t.Fatalf("%+v", got)
	}
}

func TestNormalizeLine_talkAliasToCasual(t *testing.T) {
	got := NormalizeLine(LineOut{Motion: "talk", Expression: "Neutral"})
	if got.Motion != "casual" {
		t.Fatalf("got %q", got.Motion)
	}
}

func TestNormalizeLine_happyLightSemantic(t *testing.T) {
	got := NormalizeLine(LineOut{Motion: "happy_light", Expression: "Joy"})
	if got.Motion != "happy_light" {
		t.Fatalf("got %q", got.Motion)
	}
}

func TestNormalizeLine_unknownMotion(t *testing.T) {
	got := NormalizeLine(LineOut{Motion: "NotARealMotion", Expression: "Neutral"})
	if got.Motion != defaultMotion {
		t.Fatalf("want %q got %q", defaultMotion, got.Motion)
	}
}

func TestNormalizeLine_amae2CanonicalizesToAmae(t *testing.T) {
	got := NormalizeLine(LineOut{Motion: "casual", Expression: "Amae2"})
	if got.Expression != "Amae" {
		t.Fatalf("expression: want Amae got %q", got.Expression)
	}
}

func TestNormalizeLine_legacyKissExpressionToNeutral(t *testing.T) {
	got := NormalizeLine(LineOut{Motion: "kiss", Expression: "Kiss"})
	if got.Expression != "Neutral" {
		t.Fatalf("expression: want Neutral got %q", got.Expression)
	}
}
