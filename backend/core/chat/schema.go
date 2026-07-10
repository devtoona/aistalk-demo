package chat

// 1 ブロックあたりの発話量上限（TTS 負荷・読み上げ時間の抑制。プロンプトと数を揃えること）
const (
	avatarScriptMaxLinesPerBlock = 5
	avatarScriptMaxTextChars     = 40
)

// avatarScriptLineProperties OpenAI response_format 用の lines[].* 定義（strict 用 items スキーマの properties）
func avatarScriptLineProperties() map[string]any {
	return map[string]any{
		"text": map[string]any{
			"type":        "string",
			"maxLength":   avatarScriptMaxTextChars,
			"description": "1文。上限40文字（目安は30文字前後）。無理に詰めず自然な区切りで。長い内容は次の要素へ分ける",
		},
	}
}

func avatarScriptLineItemSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"properties":           avatarScriptLineProperties(),
		"required":             []string{"text"},
		"additionalProperties": false,
	}
}

func avatarScriptLinesArraySchema() map[string]any {
	return map[string]any{
		"type":        "array",
		"minItems":    1,
		"maxItems":    avatarScriptMaxLinesPerBlock,
		"description": "話者のセグメント列。1〜5要素・空にしない。目安: 各text合計〜140文字・1行〜30文字前後。通常3〜4セグメントが理想。内容に応じ2〜5でも可。短い行に無理に詰めず自然な区切りを優先。maxItems/maxLengthは暴走防止の上限",
		"items":       avatarScriptLineItemSchema(),
	}
}

// ChatReplyJSONSchema OpenAI response_format 用（strict: true）。
// モデルは会話本文 reply と読み style のみ。TTS 用の lines はサーバが SegmentReplyToAvatarScript で組み立てる。
func ChatReplyJSONSchema() map[string]any {
	const maxReplyRunes = 3000
	return map[string]any{
		"type":        "object",
		"description": "応答はこのオブジェクトの JSON のみ。reply はこのターンの発話全文（改行可）。行区切り・セグメント分割はサーバが行う。読み上げの抑揚は style_local_id（TTS）のみ",
		"properties": map[string]any{
			"reply": map[string]any{
				"type":        "string",
				"minLength":   1,
				"maxLength":   maxReplyRunes,
				"description": "このターンのセリフ全文。句読点で息継ぎしやすい文を推奨。セグメント境界はサーバが付与する",
			},
			"style_local_id": map[string]any{
				"type":        "string",
				"pattern":     "^[0-9]+$",
				"maxLength":   8,
				"description": "読み上げスタイルID（非負整数の文字列）。この応答全体で1つ。システムプロンプトの「読み上げスタイル」一覧の local_id のみ",
			},
		},
		"required":             []string{"reply", "style_local_id"},
		"additionalProperties": false,
	}
}

// AvatarScriptJSONSchema OpenAI response_format 用（strict: true）。ルートは lines + style_local_id。話者は HTTP の space_persona_id。
// 互換・テスト用。本番チャットは ChatReplyJSONSchema を使用する。
func AvatarScriptJSONSchema() map[string]any {
	return map[string]any{
		"type":        "object",
		"description": "応答はこのオブジェクトの JSON のみ（前後に説明・マークダウン禁止）。ルートは lines と style_local_id のみ。turns・話者 id・スキーマ外キーは禁止。lines は各要素 text のみ。読み上げの抑揚・感情表現はルートの style_local_id（TTS スタイル）に任せる。モーション・表情はチャット JSON に含めない（別 API）。長い話題はこの応答では要点に留め次の user 発話に委ねる",
		"properties": map[string]any{
			"lines": avatarScriptLinesArraySchema(),
			"style_local_id": map[string]any{
				"type":        "string",
				"pattern":     "^[0-9]+$",
				"maxLength":   8,
				"description": "読み上げスタイルID（非負整数の文字列）。この応答のすべての lines で共通。システムプロンプトの「読み上げスタイル」一覧に載る local_id のみ。一覧にない数字は禁止",
			},
		},
		"required":             []string{"lines", "style_local_id"},
		"additionalProperties": false,
	}
}
