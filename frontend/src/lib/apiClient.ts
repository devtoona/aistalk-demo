import { getIdToken } from "@/lib/auth";

export class ApiError extends Error {
	status: number;
	body: string;

	constructor(status: number, body: string) {
		super(`API ${status}: ${body}`);
		this.name = "ApiError";
		this.status = status;
		this.body = body;
	}
}

/** fetch with Firebase Bearer token when configured. */
export async function authFetch(input: string, init: RequestInit = {}): Promise<Response> {
	const headers = new Headers(init.headers);
	if (!headers.has("Content-Type") && init.body) {
		headers.set("Content-Type", "application/json");
	}

	const token = await getIdToken();
	if (token) {
		headers.set("Authorization", `Bearer ${token}`);
	}

	const res = await fetch(input, { ...init, headers });
	if (res.status === 401) {
		const refreshed = await getIdToken(true);
		if (refreshed && refreshed !== token) {
			headers.set("Authorization", `Bearer ${refreshed}`);
			return fetch(input, { ...init, headers });
		}
	}
	return res;
}

/** Append access_token for EventSource (cannot set Authorization header). */
export async function withAccessToken(url: string): Promise<string> {
	const token = await getIdToken();
	if (!token) return url;
	const abs = new URL(url, typeof window !== "undefined" ? window.location.href : "http://localhost");
	abs.searchParams.set("access_token", token);
	return abs.toString();
}
