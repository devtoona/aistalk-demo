import type { AvatarScript } from "./avatarScript";
import type { Message } from "./Message";

export type { AvatarScript, AvatarScriptLine } from "./avatarScript";

export interface FetchOpenAIResponse {
	histories: Message[];
	/** アシスタントの表示用テキスト（script と整合） */
	content: string;
	script: AvatarScript;
}
