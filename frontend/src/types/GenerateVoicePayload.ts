import type { GenerateVoiceData } from "./GenerateVoiceData";

export interface GenerateVoicePayload {
	sessionId: string;
	/** Aivis 時: 選択中ペルソナの voice_masters.id（サーバで provider_payload を読む） */
	voice_master_id?: string;
	segments: GenerateVoiceData[];
}
