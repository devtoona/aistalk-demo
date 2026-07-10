package motion

import "strings"

// AllowedAnimatorMotionPhrases は Unity Animator 用のフルパス・短名（ドキュメント・フロント展開結果と揃える）。
// OpenAI の enum には含めない（意味タグは AllowedSemanticMotionTags）。
var AllowedAnimatorMotionPhrases = []string{
	"casual",
	"chatty",
	"idle",
	"thinking",
	"explain",
	"calm",
	"KA_Idle.React_SM.Angry",
	"KA_Idle.React_SM.Cry",
	"KA_Idle.React_SM.FeelDown",
	"KA_Idle.React_SM.JumpForJoy",
	"KA_Idle.React_SM.Laugh",
	"KA_Idle.React_SM.Scaring",
	"KA_Idle.React_SM.Shout",
	"KA_Idle.React_SM.Shy",
	"KA_Idle.React_SM.ShyRefusal",
	"KA_Idle.React_SM.Surprise",
	"KA_Idle.React_SM.Surprise2",
	"KA_Idle.React_SM.Surprised",
	"KA_Idle.React_SM.Taunt",
	"KA_Idle.React_SM.Tsundere",
	"KA_Idle.React_SM.Yay",
	"KA_Idle.Skinship_SM.HighFive1_1",
	"KA_Idle.Skinship_SM.Hug1_1",
	"KA_Idle.Skinship_SM.Kiss1_1",
	"KA_Idle.Pose_SM.CatPose",
	"KA_Idle.Pose_SM.CrossLegs",
	"KA_Idle.Pose_SM.CuteShyPose",
	"KA_Idle.Pose_SM.HandOnHip",
	"KA_Idle.Pose_SM.IdolPose",
	"KA_Idle.Pose_SM.Waiting",
	"KA_Idle.Greeting_SM.Cheer",
	"KA_Idle.Greeting_SM.ThumbsUp",
	"KA_Idle.Greeting_SM.Pointing",
	"KA_Idle.Greeting_SM.WaveHandSlightly",
	"KA_Idle.Exercise_SM.Backflip",
	"KA_Idle.Exercise_SM.CartwheelAndBackHandspring",
	"KA_Idle.Exercise_SM.Handstand",
	"KA_Idle.Exercise_SM.Stretch",
	"KA_Idle.Exercise_SM.Stretch2",
}

// motionAliases は OpenAI 誤出力・旧プロンプト向け。値は意味タグに寄せる。
var motionAliases = map[string]string{
	"talk": "casual",
	"Talk": "casual",
}

// canonicalMotion は旧表記を意味タグへ寄せる。Animator フルパスは意味タグへ畳む（フロントで再展開）。
func canonicalMotion(m string) string {
	s := strings.TrimSpace(m)
	if c, ok := motionAliases[s]; ok {
		s = strings.TrimSpace(c)
	}
	if strings.HasPrefix(s, "KA_Idle.") {
		return "casual"
	}
	return s
}
