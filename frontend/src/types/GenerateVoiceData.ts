export interface GenerateVoiceData {
	message: string;
	speed?: number;
	pitch?: number;
	expression?: string;
	motion?: string;
	/** script.style_local_id。VoicePeak / Aivis とも DB のスタイル一覧と突き合わせ */
	style_local_id?: string;
}
