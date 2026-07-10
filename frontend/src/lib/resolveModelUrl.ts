/**
 * VRM モデル URL を Unity 用の絶対 URL に変換する。
 * デモでは `public/models/*.vrm` を `/models/*.vrm` として同一オリジンから配信する。
 */
export function resolveModelUrl(url: string | null | undefined): string {
	const u = typeof url === "string" ? url.trim() : "";
	if (!u) return "";
	if (u.startsWith("http://") || u.startsWith("https://")) return u;
	if (typeof window !== "undefined") {
		const origin = window.location.origin;
		return `${origin}${u.startsWith("/") ? u : `/${u}`}`;
	}
	return u.startsWith("/") ? u : `/${u}`;
}
