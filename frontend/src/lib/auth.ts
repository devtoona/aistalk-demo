import { onAuthStateChanged, signInAnonymously, type User } from "firebase/auth";
import { getFirebaseAuth, isFirebaseConfigured } from "@/lib/firebase";

let readyPromise: Promise<User | null> | null = null;

/** Ensure anonymous Firebase user exists. Returns null when Firebase is not configured (local). */
export function ensureAnonymousUser(): Promise<User | null> {
	if (!isFirebaseConfigured()) {
		return Promise.resolve(null);
	}
	if (readyPromise) return readyPromise;

	readyPromise = new Promise((resolve, reject) => {
		const auth = getFirebaseAuth();
		const unsub = onAuthStateChanged(
			auth,
			async (user) => {
				unsub();
				try {
					if (user) {
						resolve(user);
						return;
					}
					const cred = await signInAnonymously(auth);
					resolve(cred.user);
				} catch (e) {
					readyPromise = null;
					reject(e);
				}
			},
			(err) => {
				readyPromise = null;
				reject(err);
			},
		);
	});

	return readyPromise;
}

export async function getIdToken(forceRefresh = false): Promise<string | null> {
	if (!isFirebaseConfigured()) return null;
	const user = await ensureAnonymousUser();
	if (!user) return null;
	return user.getIdToken(forceRefresh);
}
