package chat

import "testing"

func TestFilterSpeakableAvatarScript(t *testing.T) {
	in := AvatarScript{
		Lines: []AvatarScriptLine{
			{Text: "  hello  "},
			{Text: ""},
			{Text: "   "},
			{Text: "world"},
		},
		StyleLocalID: "1",
	}
	got := FilterSpeakableAvatarScript(in)
	if len(got.Lines) != 2 || got.Lines[0].Text != "hello" || got.Lines[1].Text != "world" || got.StyleLocalID != "1" {
		t.Fatalf("%#v", got)
	}
}
