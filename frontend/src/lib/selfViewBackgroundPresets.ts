import type { UnityRgbaPayload, UnitySelfViewBackgroundPayload } from "@/lib/unityBridge";
import { DEFAULT_SELF_VIEW_BACKGROUND } from "@/lib/unityBridge";

export const BACKGROUND_PRESET_LS_KEY = "aistalk_demo_background_preset_v1";

export type SelfViewBackgroundPreset = {
	id: string;
	label: string;
	kind: "solid" | "gradient";
	/** スウォッチ用 CSS */
	previewCss: string;
	payload: UnitySelfViewBackgroundPayload;
};

function rgba(hex: string): UnityRgbaPayload {
	const h = hex.replace("#", "");
	return {
		r: parseInt(h.slice(0, 2), 16) / 255,
		g: parseInt(h.slice(2, 4), 16) / 255,
		b: parseInt(h.slice(4, 6), 16) / 255,
		a: 1,
	};
}

function solidPreset(id: string, label: string, hex: string): SelfViewBackgroundPreset {
	const c = rgba(hex);
	return {
		id,
		label,
		kind: "solid",
		previewCss: hex,
		payload: {
			useProceduralSkybox: true,
			useGradient: false,
			main: c,
			skyboxColorTop: c,
			skyboxColorBottom: c,
			usePanorama: false,
			panoramaUrl: "",
			panoramaFlipY: false,
		},
	};
}

function gradientPreset(
	id: string,
	label: string,
	topHex: string,
	bottomHex: string,
): SelfViewBackgroundPreset {
	const top = rgba(topHex);
	const bottom = rgba(bottomHex);
	return {
		id,
		label,
		kind: "gradient",
		previewCss: `linear-gradient(180deg, ${topHex} 0%, ${bottomHex} 100%)`,
		payload: {
			useProceduralSkybox: true,
			useGradient: true,
			main: top,
			skyboxColorTop: top,
			skyboxColorBottom: bottom,
			usePanorama: false,
			panoramaUrl: "",
			panoramaFlipY: false,
		},
	};
}

export const SOLID_BACKGROUND_PRESETS: SelfViewBackgroundPreset[] = [
	solidPreset("cream", "クリーム", "#FDF1EF"),
	solidPreset("white", "ホワイト", "#FFFFFF"),
	solidPreset("sky", "スカイ", "#E8F4FC"),
	solidPreset("lavender", "ラベンダー", "#EDE8F5"),
	solidPreset("mint", "ミント", "#E8F5F0"),
	solidPreset("peach", "ピーチ", "#FFE8E0"),
	solidPreset("sand", "サンド", "#F5EFE6"),
	solidPreset("slate", "スレート", "#2D3748"),
];

export const GRADIENT_BACKGROUND_PRESETS: SelfViewBackgroundPreset[] = [
	gradientPreset("starry-sky", "星空", "#1B2A4A", "#060B14"),
	gradientPreset("sunrise", "サンライズ", "#FFE8D6", "#FFD1DC"),
	gradientPreset("day-sky", "昼空", "#B8E4FF", "#FDF1EF"),
	gradientPreset("sakura", "桜", "#FFD6E8", "#FFF5F7"),
	gradientPreset("twilight", "トワイライト", "#C4B5FD", "#312E81"),
	gradientPreset("ocean", "オーシャン", "#67E8F9", "#1E3A8A"),
	gradientPreset("sunset", "サンセット", "#FDE68A", "#F97316"),
	gradientPreset("aurora", "オーロラ", "#A7F3D0", "#6366F1"),
	gradientPreset("night", "ナイト", "#1E293B", "#0F172A"),
];

export const ALL_BACKGROUND_PRESETS: SelfViewBackgroundPreset[] = [
	...SOLID_BACKGROUND_PRESETS,
	...GRADIENT_BACKGROUND_PRESETS,
];

export function getDefaultBackgroundPreset(): SelfViewBackgroundPreset {
	return (
		findBackgroundPresetById("starry-sky") ?? {
			id: "default",
			label: "星空",
			kind: "gradient",
			previewCss: "linear-gradient(180deg, #1B2A4A 0%, #060B14 100%)",
			payload: GRADIENT_BACKGROUND_PRESETS[0]?.payload ?? DEFAULT_SELF_VIEW_BACKGROUND,
		}
	);
}

export function findBackgroundPresetById(id: string): SelfViewBackgroundPreset | undefined {
	return ALL_BACKGROUND_PRESETS.find((p) => p.id === id);
}

export function loadBackgroundPresetIdFromStorage(): string | null {
	if (typeof window === "undefined") return null;
	try {
		const raw = localStorage.getItem(BACKGROUND_PRESET_LS_KEY);
		return raw?.trim() || null;
	} catch {
		return null;
	}
}

export function saveBackgroundPresetIdToStorage(id: string) {
	if (typeof window === "undefined") return;
	try {
		localStorage.setItem(BACKGROUND_PRESET_LS_KEY, id);
	} catch {
		// ignore
	}
}

export function resolveBackgroundPreset(id?: string | null): SelfViewBackgroundPreset {
	if (id) {
		const found = findBackgroundPresetById(id);
		if (found) return found;
	}
	const stored = loadBackgroundPresetIdFromStorage();
	if (stored) {
		const found = findBackgroundPresetById(stored);
		if (found) return found;
	}
	return getDefaultBackgroundPreset();
}
