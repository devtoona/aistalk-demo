package motion

// AllowedSemanticMotionTags は OpenAI strict JSON の motion enum と一致（フロント motionSemanticExpand と揃える）。
var AllowedSemanticMotionTags = []string{
	"idle",
	"casual",
	"explain",
	"calm",
	"thinking",
	"happy_light",
	"happy_big",
	"shy",
	"tsundere",
	"sad",
	"angry",
	"surprise",
	"greeting",
	"pose",
	"hug",
	"kiss",
	"highfive",
}

var allowedSemanticMotionSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(AllowedSemanticMotionTags))
	for _, s := range AllowedSemanticMotionTags {
		m[s] = struct{}{}
	}
	return m
}()
