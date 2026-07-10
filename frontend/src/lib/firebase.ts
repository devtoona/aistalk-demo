import { initializeApp, getApps, type FirebaseApp } from "firebase/app";
import { getAuth, type Auth } from "firebase/auth";

function requiredEnv(name: string): string {
	const v = process.env[name]?.trim();
	if (!v) {
		throw new Error(`Missing ${name}`);
	}
	return v;
}

let app: FirebaseApp | null = null;
let auth: Auth | null = null;

export function isFirebaseConfigured(): boolean {
	return Boolean(process.env.NEXT_PUBLIC_FIREBASE_API_KEY?.trim());
}

export function getFirebaseAuth(): Auth {
	if (auth) return auth;
	if (!getApps().length) {
		app = initializeApp({
			apiKey: requiredEnv("NEXT_PUBLIC_FIREBASE_API_KEY"),
			authDomain: requiredEnv("NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN"),
			projectId: requiredEnv("NEXT_PUBLIC_FIREBASE_PROJECT_ID"),
			appId: requiredEnv("NEXT_PUBLIC_FIREBASE_APP_ID"),
		});
	} else {
		app = getApps()[0]!;
	}
	auth = getAuth(app);
	return auth;
}
