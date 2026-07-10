export type DemoSelfAvatar = {
	id: string;
	spaceId: string;
	personaId: string;
	label: string;
	modelUrl: string;
};

export const ASSISTANT_DISPLAY_LABEL = "アシスタント";

export type DemoPersona = {
	id: string;
	spaceId: string;
	personaId: string;
	nickname: string;
	personaName: string;
	modelUrl: string;
	voiceMasterId: string;
	personality: string;
	responseStyle: string;
};

const DEFAULT_SPACE_ID = "demo-space-1";

const DEFAULT_SELF_AVATAR: DemoSelfAvatar = {
	id: "demo-self-1",
	spaceId: DEFAULT_SPACE_ID,
	personaId: "demo-self-1",
	label: "あなた",
	modelUrl: "/models/self.vrm",
};

const DEFAULT_OWNED_PERSONA: DemoPersona = {
	id: "demo-persona-1",
	spaceId: DEFAULT_SPACE_ID,
	personaId: "demo-persona-1",
	nickname: ASSISTANT_DISPLAY_LABEL,
	personaName: ASSISTANT_DISPLAY_LABEL,
	modelUrl: "/models/avatar.vrm",
	voiceMasterId: "",
	personality: "明るく丁寧。短めに話す。",
	responseStyle: "共感的",
};

export const MESSAGES_LS_KEY = "aistalk_demo_messages_v1";

export function getDemoSelfAvatar(): DemoSelfAvatar {
	return {
		...DEFAULT_SELF_AVATAR,
		label: process.env.NEXT_PUBLIC_SELF_LABEL?.trim() || DEFAULT_SELF_AVATAR.label,
		modelUrl: process.env.NEXT_PUBLIC_SELF_VRM_URL?.trim() || DEFAULT_SELF_AVATAR.modelUrl,
	};
}

export function getDemoPersona(): DemoPersona {
	const envName = process.env.NEXT_PUBLIC_PERSONA_NAME?.trim() || "";
	return {
		...DEFAULT_OWNED_PERSONA,
		nickname: envName || DEFAULT_OWNED_PERSONA.nickname,
		personaName: envName || DEFAULT_OWNED_PERSONA.personaName,
		modelUrl: process.env.NEXT_PUBLIC_VRM_URL?.trim() || DEFAULT_OWNED_PERSONA.modelUrl,
		voiceMasterId: process.env.NEXT_PUBLIC_AIVIS_MODEL_UUID?.trim() || "",
		personality: process.env.NEXT_PUBLIC_PERSONA_PERSONALITY?.trim() || DEFAULT_OWNED_PERSONA.personality,
		responseStyle: process.env.NEXT_PUBLIC_PERSONA_RESPONSE_STYLE?.trim() || DEFAULT_OWNED_PERSONA.responseStyle,
	};
}
