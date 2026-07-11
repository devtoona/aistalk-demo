import { initializeApp, getApps, type FirebaseApp } from "firebase/app";
import { getAuth, type Auth } from "firebase/auth";

/**
 * Next.js は `process.env.NEXT_PUBLIC_*` の静的参照だけをビルド時に置換する。
 * `process.env[name]` のような動的アクセスはクライアントバンドルで常に undefined になる。
 */
export function isFirebaseConfigured(): boolean {
	return Boolean(
		process.env.NEXT_PUBLIC_FIREBASE_API_KEY?.trim() &&
			process.env.NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN?.trim() &&
			process.env.NEXT_PUBLIC_FIREBASE_PROJECT_ID?.trim() &&
			process.env.NEXT_PUBLIC_FIREBASE_APP_ID?.trim(),
	);
}

let app: FirebaseApp | null = null;
let auth: Auth | null = null;

export function getFirebaseAuth(): Auth {
	if (auth) return auth;

	const apiKey = process.env.NEXT_PUBLIC_FIREBASE_API_KEY?.trim() ?? "";
	const authDomain = process.env.NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN?.trim() ?? "";
	const projectId = process.env.NEXT_PUBLIC_FIREBASE_PROJECT_ID?.trim() ?? "";
	const appId = process.env.NEXT_PUBLIC_FIREBASE_APP_ID?.trim() ?? "";

	const missing: string[] = [];
	if (!apiKey) missing.push("NEXT_PUBLIC_FIREBASE_API_KEY");
	if (!authDomain) missing.push("NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN");
	if (!projectId) missing.push("NEXT_PUBLIC_FIREBASE_PROJECT_ID");
	if (!appId) missing.push("NEXT_PUBLIC_FIREBASE_APP_ID");
	if (missing.length > 0) {
		throw new Error(`Missing ${missing.join(", ")}`);
	}

	if (!getApps().length) {
		app = initializeApp({ apiKey, authDomain, projectId, appId });
	} else {
		app = getApps()[0]!;
	}
	auth = getAuth(app);
	return auth;
}
