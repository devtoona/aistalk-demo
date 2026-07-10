/** サーバが最新 user のみ付与するオブジェクト（OpenAI 向け） */
export type ConversationSituationDetail = {
	turn_candidates: string[];
	situation_summary: string;
};

/** POST/GET /api/chat の messages[].content と同型（JSON 文字列として送受信） */
export type ChatTurnContent = {
	id: string;
	name: string;
	message: string;
	/** サーバ注入。文字列は旧形式の要約のみ */
	conversation_situation?: string | ConversationSituationDetail;
};

export function stringifyChatTurnContent(c: ChatTurnContent): string {
	const o: Record<string, unknown> = {
		id: c.id ?? "",
		name: c.name ?? "",
		message: c.message ?? "",
	};
	if (c.conversation_situation !== undefined && c.conversation_situation !== null) {
		o.conversation_situation = c.conversation_situation;
	}
	return JSON.stringify(o);
}

/** 表示・整形用。JSON でなければ全体を message とみなす */
export function parseChatTurnContent(raw: string): ChatTurnContent {
	const s = raw.trim();
	if (!s.startsWith("{")) {
		return { id: "", name: "", message: raw };
	}
	try {
		const o = JSON.parse(s) as Record<string, unknown>;
		const sit = o.conversation_situation;
		let conversation_situation: string | ConversationSituationDetail | undefined;
		if (typeof sit === "string") {
			conversation_situation = sit;
		} else if (sit && typeof sit === "object" && !Array.isArray(sit)) {
			const d = sit as Record<string, unknown>;
			const tc = d.turn_candidates;
			conversation_situation = {
				turn_candidates: Array.isArray(tc) ? tc.filter((x): x is string => typeof x === "string") : [],
				situation_summary: typeof d.situation_summary === "string" ? d.situation_summary : "",
			};
		}
		return {
			id: typeof o.id === "string" ? o.id : "",
			name: typeof o.name === "string" ? o.name : "",
			message: typeof o.message === "string" ? o.message : "",
			conversation_situation,
		};
	} catch {
		return { id: "", name: "", message: raw };
	}
}

/** assistant の content は `[{id,name,message},…]`。user や旧単一オブジェクトは 1 要素に正規化する。 */
export function parseChatTurnSegments(raw: string): ChatTurnContent[] {
	const s = raw.trim();
	if (!s) {
		return [{ id: "", name: "", message: "" }];
	}
	if (s.startsWith("[")) {
		try {
			const arr = JSON.parse(s) as unknown[];
			if (!Array.isArray(arr)) {
				return [{ id: "", name: "", message: raw }];
			}
			return arr.map((item) => {
				if (item && typeof item === "object" && !Array.isArray(item)) {
					const o = item as Record<string, unknown>;
					return {
						id: typeof o.id === "string" ? o.id : "",
						name: typeof o.name === "string" ? o.name : "",
						message: typeof o.message === "string" ? o.message : "",
					};
				}
				return { id: "", name: "", message: "" };
			});
		} catch {
			return [{ id: "", name: "", message: raw }];
		}
	}
	return [parseChatTurnContent(raw)];
}

/** 履歴 UI 用。assistant の `[{id,name,message}]` も message 本文だけに展開する。 */
export function formatChatMessageForDisplay(
	raw: string,
	fallbackName = "",
): { name: string; message: string } {
	const segs = parseChatTurnSegments(raw);
	const message =
		segs
			.map((x) => x.message.trim())
			.filter(Boolean)
			.join("\n") || raw.trim();
	const name = segs.map((x) => x.name.trim()).find(Boolean) || fallbackName;
	return { name, message };
}

/** format API 用に、末尾から見て最初の user / assistant を返す（system 等はスキップ）。 */
export function lastUserOrAssistantForFormat(messages: { role: string; content: string }[]): {
	role: "user" | "assistant";
	content: string;
} | undefined {
	for (let i = messages.length - 1; i >= 0; i--) {
		const m = messages[i];
		if (m.role === "user" || m.role === "assistant") {
			return { role: m.role, content: m.content };
		}
	}
	return undefined;
}

/** 直近メッセージを format API 用に。assistant は配列要素の message を改行連結。 */
export function lastMessageTextForFormat(raw: string, role: "user" | "assistant"): string | undefined {
	if (role === "user") {
		const t = parseChatTurnContent(raw).message.trim();
		return t || undefined;
	}
	const segs = parseChatTurnSegments(raw);
	const t = segs
		.map((x) => x.message.trim())
		.filter(Boolean)
		.join("\n")
		.trim();
	return t || undefined;
}
