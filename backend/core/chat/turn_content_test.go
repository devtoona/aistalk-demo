package chat

import "testing"

func TestNormalizeMessageContentForModel_UserPlain(t *testing.T) {
	got := NormalizeMessageContentForModel("user", "やあ", "hoge", "太郎")
	if got != "やあ" {
		t.Fatalf("got %q want plain utterance", got)
	}
}

func TestNormalizeMessageContentForModel_UserJSON(t *testing.T) {
	raw := `{"id":"hoge","name":"太郎","message":"やあ"}`
	got := NormalizeMessageContentForModel("user", raw, "x", "y")
	if got != "やあ" {
		t.Fatalf("extract message: got %q", got)
	}
}

func TestNormalizeMessageContentForModel_AssistantPlain(t *testing.T) {
	got := NormalizeMessageContentForModel("assistant", "元気？", "", "")
	if got != `[{"id":"","name":"","message":"元気？"}]` {
		t.Fatalf("got %q", got)
	}
}

func TestNormalizeMessageContentForModel_AssistantLegacySingleObject(t *testing.T) {
	raw := `{"id":"npc-1","name":"B","message":"hello"}`
	got := NormalizeMessageContentForModel("assistant", raw, "", "")
	want := `[{"id":"npc-1","name":"B","message":"hello"}]`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestMarshalChatTurnJSON(t *testing.T) {
	s, err := MarshalChatTurnJSON("fuga", "二郎", "元気？")
	if err != nil {
		t.Fatal(err)
	}
	if s != `{"id":"fuga","name":"二郎","message":"元気？"}` {
		t.Fatalf("got %q", s)
	}
}
