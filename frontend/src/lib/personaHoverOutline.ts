import { sendPersonaHoverOutlineVisibleToUnity } from "@/lib/unityBridge";

export const PERSONA_HOVER_OUTLINE_VISIBLE_LS_KEY = "aistalk_demo_hover_outline_visible_v2";

/** 初回・未保存時は OFF */
export function loadPersonaHoverOutlineVisible(): boolean {
	if (typeof window === "undefined") return false;
	try {
		const raw = localStorage.getItem(PERSONA_HOVER_OUTLINE_VISIBLE_LS_KEY);
		if (raw === null) return false;
		return raw === "1";
	} catch {
		return false;
	}
}

export function savePersonaHoverOutlineVisible(visible: boolean): void {
	if (typeof window === "undefined") return;
	try {
		localStorage.setItem(PERSONA_HOVER_OUTLINE_VISIBLE_LS_KEY, visible ? "1" : "0");
	} catch {
		// ignore quota errors
	}
}

export function applyPersonaHoverOutlineVisible(visible: boolean): boolean {
	savePersonaHoverOutlineVisible(visible);
	return sendPersonaHoverOutlineVisibleToUnity(visible);
}
