/**
 * モーション API は意味タグのみ返す。Unity 向け具体名への展開はこのモジュールに集約する。
 * プール選択は segmentIndex とタグから決定的（再現しやすい）。
 */

export const SEMANTIC_MOTION_TAGS = [
	"idle",
	"casual",
	"explain",
	"calm",
	"thinking",
	"happy_light",
	"happy_big",
	"shy",
	"tsundere",
	"sad",
	"angry",
	"surprise",
	"greeting",
	"pose",
	"hug",
	"kiss",
	"highfive",
] as const;

export type SemanticMotionTag = (typeof SEMANTIC_MOTION_TAGS)[number];

const SEMANTIC_SET: ReadonlySet<string> = new Set(SEMANTIC_MOTION_TAGS);

/** OpenAI / バックエンドが返す意味タグか */
export function isSemanticMotionTag(motion: string): boolean {
	return SEMANTIC_SET.has(motion.trim());
}

function fnv1aPick(tag: string, segmentIndex: number, modulo: number): number {
	if (modulo <= 0) return 0;
	let h = 2166136261;
	const s = `${tag}\0${segmentIndex}`;
	for (let i = 0; i < s.length; i++) {
		h ^= s.charCodeAt(i);
		h = Math.imul(h, 16777619);
	}
	return Math.abs(h) % modulo;
}

function pickPool(tag: string, segmentIndex: number, pool: readonly string[]): string {
	if (pool.length === 0) return "idle";
	if (pool.length === 1) return pool[0]!;
	const i = fnv1aPick(tag, segmentIndex, pool.length);
	return pool[i]!;
}

/** Unity MotionController / Animator と一致するフルパスまたは既存短名 */
const POOLS: Record<SemanticMotionTag, readonly string[]> = {
	idle: ["idle"],
	casual: ["casual", "chatty"],
	explain: ["explain"],
	calm: ["calm"],
	thinking: ["thinking"],
	happy_light: ["KA_Idle.React_SM.Yay", "KA_Idle.React_SM.Laugh"],
	happy_big: ["KA_Idle.React_SM.JumpForJoy"],
	shy: ["KA_Idle.React_SM.Shy", "KA_Idle.Pose_SM.CuteShyPose"],
	tsundere: ["KA_Idle.React_SM.Tsundere", "KA_Idle.React_SM.ShyRefusal"],
	sad: ["KA_Idle.React_SM.Cry", "KA_Idle.React_SM.FeelDown"],
	angry: ["KA_Idle.React_SM.Angry", "KA_Idle.React_SM.Taunt"],
	surprise: ["KA_Idle.React_SM.Surprise", "KA_Idle.React_SM.Surprise2", "KA_Idle.React_SM.Surprised"],
	greeting: [
		"KA_Idle.Greeting_SM.WaveHandSlightly",
		"KA_Idle.Greeting_SM.Cheer",
		"KA_Idle.Greeting_SM.ThumbsUp",
	],
	pose: [
		"KA_Idle.Pose_SM.CatPose",
		"KA_Idle.Pose_SM.IdolPose",
		"KA_Idle.Pose_SM.HandOnHip",
		"KA_Idle.Pose_SM.CrossLegs",
	],
	hug: ["KA_Idle.Skinship_SM.Hug1_1"],
	kiss: ["KA_Idle.Skinship_SM.Kiss1_1"],
	highfive: ["KA_Idle.Skinship_SM.HighFive1_1"],
};

/**
 * 意味タグを Unity に送るモーション名に変換する。
 * 既に KA_Idle. で始まる場合はそのまま返す（デバッグ・後方互換）。
 * 未知の文字列は casual プールとして扱う。
 */
export function expandSemanticMotion(motion: string, segmentIndex: number): string {
	const raw = motion.trim();
	if (raw.startsWith("KA_Idle.")) {
		return raw;
	}
	const tag = (isSemanticMotionTag(raw) ? raw : "casual") as SemanticMotionTag;
	const pool = POOLS[tag];
	return pickPool(tag, segmentIndex, pool);
}

/** Unity 送信用: 意味タグなら展開、Animator フルパスはそのまま、それ以外（Nod 等）はそのまま */
export function resolvePlayMotionName(name: string, segmentIndex: number): string {
	const t = name.trim();
	if (!t) {
		return expandSemanticMotion("casual", segmentIndex);
	}
	if (t.startsWith("KA_Idle.")) {
		return t;
	}
	if (isSemanticMotionTag(t)) {
		return expandSemanticMotion(t, segmentIndex);
	}
	return t;
}

/** 開発者向け: 各タグ×複数 index で展開結果のユニーク一覧（Unity 手動試験用） */
export function sampleExpandedMotions(maxIndexPerTag = 16): string[] {
	const set = new Set<string>();
	for (const tag of SEMANTIC_MOTION_TAGS) {
		for (let i = 0; i < maxIndexPerTag; i++) {
			set.add(expandSemanticMotion(tag, i));
		}
	}
	return Array.from(set);
}
