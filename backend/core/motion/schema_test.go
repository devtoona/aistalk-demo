package motion

import "testing"

func TestMotionScriptJSONSchemaShape(t *testing.T) {
	s := MotionScriptJSONSchema()
	props, _ := s["properties"].(map[string]any)
	if props["lines"] == nil {
		t.Fatal("missing lines")
	}
	req, _ := s["required"].([]string)
	if len(req) != 1 || req[0] != "lines" {
		t.Fatalf("root required: %#v", req)
	}
}
