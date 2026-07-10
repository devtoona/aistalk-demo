package chat

import (
	"encoding/json"
	"testing"
)

func TestAvatarScriptJSONSchemaShape(t *testing.T) {
	s := AvatarScriptJSONSchema()
	if s["type"] != "object" {
		t.Fatalf("root type: %v", s["type"])
	}
	props, _ := s["properties"].(map[string]any)
	if props == nil {
		t.Fatal("missing properties")
	}
	if props["lines"] == nil {
		t.Fatal("missing lines")
	}
	if props["style_local_id"] == nil {
		t.Fatal("missing style_local_id")
	}
	req, _ := s["required"].([]string)
	if len(req) != 2 {
		t.Fatalf("root required len: %#v", req)
	}
	linesProp, _ := props["lines"].(map[string]any)
	items, _ := linesProp["items"].(map[string]any)
	lineReq, _ := items["required"].([]string)
	if len(lineReq) != 1 {
		t.Fatalf("lines item required len: %#v", lineReq)
	}
	if linesProp["maxItems"] != float64(5) && linesProp["maxItems"] != 5 {
		t.Fatalf("lines maxItems: want 5 got %#v", linesProp["maxItems"])
	}
	textProp, _ := items["properties"].(map[string]any)["text"].(map[string]any)
	if textProp == nil {
		t.Fatal("missing lines.items.properties.text")
	}
	if textProp["maxLength"] != float64(40) && textProp["maxLength"] != 40 {
		t.Fatalf("text maxLength: want 40 got %#v", textProp["maxLength"])
	}
}

func TestChatReplyJSONSchemaShape(t *testing.T) {
	s := ChatReplyJSONSchema()
	if s["type"] != "object" {
		t.Fatalf("root type: %v", s["type"])
	}
	props, _ := s["properties"].(map[string]any)
	if props == nil || props["reply"] == nil || props["style_local_id"] == nil {
		t.Fatalf("properties: %#v", props)
	}
	req, _ := s["required"].([]string)
	if len(req) != 2 {
		t.Fatalf("required: %#v", req)
	}
}

func TestAvatarScriptJSONRoundTrip(t *testing.T) {
	in := AvatarScript{
		Lines: []AvatarScriptLine{{Text: "hello"}},
		StyleLocalID: "5",
	}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out AvatarScript
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Lines) != 1 || out.Lines[0].Text != "hello" || out.StyleLocalID != "5" {
		t.Fatalf("unmarshal: %#v", out)
	}
}
