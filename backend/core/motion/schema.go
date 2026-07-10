package motion

// チャット API と同じ上限（プロンプト・検証を揃える）
const (
	MaxLinesPerRequest = 10
	// 従来: 1 リクエスト行あたり（チャットの script 行に揃える）
	MaxTextChars = 200
	// delegate_segmentation: 1 要素に載せるアシスタント全文の上限
	MaxDelegateAssistantRunes = 4000
	// 出力 lines[i].text の 1 行上限（TTS 1 チャンク想定）
	MaxMotionLineTextRunes = 600
)

func motionLineOutputItemSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"text": map[string]any{
				"type":        "string",
				"minLength":   1,
				"maxLength":   MaxMotionLineTextRunes,
				"description": "このセグメントの読み上げテキスト。delegate_segmentation=false のときは入力 segments[i].text と完全一致。true のときは全文を意味が通る単位で分割した各行",
			},
			"motion": map[string]any{
				"type":        "string",
				"enum":        AllowedSemanticMotionTags,
				"description": "意味タグのみ（プロンプトの一覧と完全一致）。Unity 用の具体モーションはクライアントが展開する",
			},
			"expression": map[string]any{
				"type": "string",
				"enum": []string{
					"Neutral", "Happiness", "Sad", "Angry", "Surprised",
					"Joy", "None",
					// 細かい演技用（Unity ExpressionController のプリセット名と一致）
					"Tsundere", "Hanikami", "KomariDere", "Saiaku", "Nemui", "Koakuma", "Jitome",
					// ブレンドシェープ直指定（フロント facialPreset と public/expression と一致）
					"Amae",
				},
				"description": "表情（Tsundere=ツンデレ系、Hanikami=はにかみ照れ笑い、KomariDere=困りデレ、Saiaku=最悪系、Nemui=眠そう、Koakuma=小悪魔、Jitome=ジト目、Amae=甘える※クライアントが複数フェイシャルからランダム選択。キス顔は motion kiss 再生時のみクライアントが付与）",
			},
		},
		"required":             []string{"text", "motion", "expression"},
		"additionalProperties": false,
	}
}

// MotionScriptJSONSchema OpenAI response_format 用（strict: true）。
func MotionScriptJSONSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"lines": map[string]any{
				"type":        "array",
				"minItems":    1,
				"maxItems":    MaxLinesPerRequest,
				"description": "各要素に text（読み上げ文）・motion・expression。delegate_segmentation=false のときは入力と同じ件数で text は入力と同一。true のときは 1〜10 行に再分割",
				"items":       motionLineOutputItemSchema(),
			},
		},
		"required":             []string{"lines"},
		"additionalProperties": false,
	}
}
