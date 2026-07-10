import { authFetch } from "@/lib/apiClient";
import { isSemanticMotionTag } from "@/lib/motionSemanticExpand";

const API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL || "";

export type AvatarMotionLine = {
	text?: string;
	motion: string;
	expression: string;
};

const MAX_MOTION_LINES = 10;

export type AvatarMotionInferRequest = {
	lines: { text: string }[];
	delegate_segmentation?: boolean;
	style_local_id?: string;
	last_user_message?: string;
};

export type AvatarMotionInferResponse = {
	lines: AvatarMotionLine[];
	used_fallback?: boolean;
};

/** POST /api/avatar/motion。失敗時は casual + Neutral で n 件埋める。 */
export async function inferAvatarMotion(
	body: AvatarMotionInferRequest,
): Promise<AvatarMotionInferResponse> {
	const n = body.lines.length;
	const delegate = Boolean(body.delegate_segmentation);
	const fallback = (): AvatarMotionInferResponse => {
		if (delegate && n >= 1) {
			const full = body.lines[0]?.text?.trim() ?? "";
			return {
				lines: full
					? [{ text: full, motion: "casual", expression: "Neutral" }]
					: [{ motion: "casual", expression: "Neutral" }],
				used_fallback: true,
			};
		}
		return {
			lines: Array.from({ length: n }, () => ({ motion: "casual", expression: "Neutral" })),
			used_fallback: true,
		};
	};
	if (n === 0) {
		return { lines: [], used_fallback: true };
	}
	try {
		const res = await authFetch(`${API_BASE}/api/avatar/motion`, {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify(body),
		});
		if (!res.ok) {
			console.warn("avatar/motion HTTP", res.status, await res.text().catch(() => ""));
			return fallback();
		}
		const data = (await res.json()) as AvatarMotionInferResponse;
		if (!Array.isArray(data.lines)) {
			console.warn("avatar/motion missing lines");
			return fallback();
		}
		if (delegate) {
			if (data.lines.length < 1 || data.lines.length > MAX_MOTION_LINES) {
				console.warn("avatar/motion delegate line count out of range");
				return fallback();
			}
			const bad = data.lines.some(
				(ln) =>
					typeof ln.motion !== "string" ||
					typeof ln.expression !== "string" ||
					typeof ln.text !== "string" ||
					!ln.text.trim(),
			);
			if (bad) {
				console.warn("avatar/motion delegate line missing text or motion");
				return fallback();
			}
		} else if (data.lines.length !== n) {
			console.warn("avatar/motion line count mismatch");
			return fallback();
		}
		const badSemantic = data.lines.some(
			(ln) => typeof ln.motion !== "string" || !isSemanticMotionTag(ln.motion.trim()),
		);
		if (badSemantic) {
			console.warn("avatar/motion invalid semantic motion tag");
			return fallback();
		}
		return data;
	} catch (e) {
		console.warn("avatar/motion fetch failed", e);
		return fallback();
	}
}
