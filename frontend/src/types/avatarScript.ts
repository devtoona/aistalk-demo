import type { GenerateVoiceData } from "./GenerateVoiceData";

/** POST /api/chat の script（lines + style_local_id）。話者はリクエストの space_persona_id。 */

export interface AvatarScriptLine {
	text: string;
	/** TTS スタイル（数字文字列） */
	style_local_id?: string;
}

export interface AvatarScript {
	lines: AvatarScriptLine[];
	/** この応答全体の読み上げスタイル（全セグメント共通） */
	style_local_id?: string;
}

export function avatarScriptLinesToGenerateVoiceData(script: AvatarScript): GenerateVoiceData[] {
	const defaultSid = (script.style_local_id ?? "").trim();
	const out: GenerateVoiceData[] = [];
	for (const line of script.lines) {
		const message = line.text.trim();
		if (!message) continue;
		const sid = (line.style_local_id ?? defaultSid ?? "").trim();
		out.push({
			message,
			...(sid !== "" ? { style_local_id: sid } : {}),
		});
	}
	return out;
}
