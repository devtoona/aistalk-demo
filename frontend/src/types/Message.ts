import type { AvatarScriptLine } from "./avatarScript";

/** POST /api/chat の messages[]。content は JSON 文字列 { id, name, message [, conversation_situation ] }（situation はオブジェクトまたは旧来の文字列）。lines は任意。 */
export interface Message {
	/** DB 永続行の id（GET /chat/messages）。未設定時は createdAt をキーにする */
	id?: string;
	/** DB の挿入順（スペース横断タイムライン用） */
	orderId?: number;
	role: "user" | "assistant" | "system";
	content: string;
	createdAt: string;
	lines?: AvatarScriptLine[];
	/** スペース全体表示時: スレッドの表示名（persona_settings.name 等） */
	threadLabel?: string;
	spacePersonaId?: string;
}
