import type { FacialPresetJson } from "./unityBridge";

/**
 * モーション API の expression 名 → Developer が押すのと同じ public/expression/*.json の内容。
 * （ExpressionController の Koakuma 等プリセットとは別物）
 */
const KOAKUMA: FacialPresetJson = {
	situation: "小悪魔（へぇ〜そんなことまでしちゃうんだ〜的なの）",
	parameters: [
		{ Name: "Fcl_BRW_Fun", Value: 0.25999999046325686 },
		{ Name: "Fcl_EYE_Sorrow", Value: 1.0 },
		{ Name: "Fcl_MTH_Angry", Value: 0.15000000596046449 },
		{ Name: "Fcl_MTH_Small", Value: 0.28999999165534975 },
		{ Name: "Fcl_MTH_Fun", Value: 0.4300000071525574 },
	],
};

const NEMUI: FacialPresetJson = {
	situation: "眠い",
	parameters: [
		{ Name: "Fcl_BRW_Angry", Value: 0.12999999523162843 },
		{ Name: "Fcl_BRW_Joy", Value: 0.1599999964237213 },
		{ Name: "Fcl_BRW_Surprised", Value: 0.05000000074505806 },
		{ Name: "Fcl_EYE_Fun", Value: 0.800000011920929 },
		{ Name: "Fcl_EYE_Sorrow", Value: 1.0 },
		{ Name: "Fcl_EYE_Highlight_Hide", Value: 0.5799999833106995 },
		{ Name: "Fcl_MTH_Angry", Value: 0.17000000178813935 },
	],
};

const JITOME: FacialPresetJson = {
	situation: "ジト目(退屈）",
	parameters: [
		{ Name: "Fcl_BRW_Fun", Value: 0.25999999046325686 },
		{ Name: "Fcl_EYE_Sorrow", Value: 1.0 },
		{ Name: "Fcl_MTH_Angry", Value: 0.15000000596046449 },
		{ Name: "Fcl_MTH_Small", Value: 0.28999999165534975 },
	],
};

const SAIAKU: FacialPresetJson = {
	situation: "最悪",
	parameters: [
		{ Name: "Fcl_BRW_Angry", Value: 0.5630000233650208 },
		{ Name: "Fcl_BRW_Fun", Value: 0.22699999809265138 },
		{ Name: "Fcl_BRW_Joy", Value: 0.1599999964237213 },
		{ Name: "Fcl_EYE_Sorrow", Value: 1.0 },
		{ Name: "Fcl_EYE_Spread", Value: 0.1340000033378601 },
		{ Name: "Fcl_EYE_Highlight_Hide", Value: 0.5799999833106995 },
		{ Name: "Fcl_MTH_Angry", Value: 0.28600001335144045 },
	],
};

/** 甘える（public/expression/amae.json と同一） */
const AMAE: FacialPresetJson = {
	situation: "甘える",
	parameters: [
		{ Name: "Fcl_BRW_Joy", Value: 0.5189999938011169 },
		{ Name: "Fcl_BRW_Surprised", Value: 0.5070000290870667 },
		{ Name: "Fcl_EYE_Fun", Value: 0.4390000104904175 },
		{ Name: "Fcl_MTH_Small", Value: 0.3179999887943268 },
	],
	blush: 0.65,
};

/** 甘える2（目元ジョイ弱め。public/expression/amae2.json と同一） */
const AMAE2: FacialPresetJson = {
	situation: "甘える2",
	parameters: [
		{ Name: "Fcl_BRW_Joy", Value: 0.5189999938011169 },
		{ Name: "Fcl_BRW_Surprised", Value: 0.5070000290870667 },
		{ Name: "Fcl_EYE_Fun", Value: 0.4390000104904175 },
		{ Name: "Fcl_EYE_Joy_R", Value: 0.19599999487400056 },
		{ Name: "Fcl_EYE_Joy_L", Value: 0.19599999487400056 },
		{ Name: "Fcl_MTH_Small", Value: 0.3179999887943268 },
	],
	blush: 0.65,
};

/**
 * motion kiss 再生時のみ使用（モーション API の expression 候補には含めない）。
 * public/expression/kiss.json と同一。
 */
export const KISS_MOTION_FACE_PRESET: FacialPresetJson = {
	situation: "キス",
	parameters: [
		{ Name: "Fcl_BRW_Fun", Value: 0.0820000022649765 },
		{ Name: "Fcl_BRW_Joy", Value: 1.0 },
		{ Name: "Fcl_BRW_Surprised", Value: 0.17399999499320985 },
		{ Name: "Fcl_EYE_Joy_R", Value: 1.0 },
		{ Name: "Fcl_EYE_Joy_L", Value: 1.0 },
		{ Name: "Fcl_MTH_Small", Value: 0.3179999887943268 },
	],
	blush: 0.75,
};

const BY_NORM: Record<string, FacialPresetJson> = {
	koakuma: KOAKUMA,
	nemui: NEMUI,
	jitome: JITOME,
	saiaku: SAIAKU,
};

/** API は "Amae" のみ。Unity 送信時に amae.json / amae2.json 相当を 50% で選ぶ */
export function randomAmaeFacialPreset(): FacialPresetJson {
	return Math.random() < 0.5 ? AMAE : AMAE2;
}

/** 小文字正規化名（例: koakuma）で引く（Amae は randomAmaeFacialPreset を使う） */
export function facialPresetForMotionExpression(norm: string): FacialPresetJson | undefined {
	return BY_NORM[norm];
}
