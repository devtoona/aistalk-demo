"use client";

import { createContext, useContext, useEffect, useState, type ReactNode } from "react";
import { ensureAnonymousUser } from "@/lib/auth";
import { isFirebaseConfigured } from "@/lib/firebase";

type AuthState = {
	ready: boolean;
	error: string | null;
};

const AuthContext = createContext<AuthState>({ ready: false, error: null });

export function AuthProvider({ children }: { children: ReactNode }) {
	const [ready, setReady] = useState(!isFirebaseConfigured());
	const [error, setError] = useState<string | null>(null);

	useEffect(() => {
		if (!isFirebaseConfigured()) {
			setReady(true);
			return;
		}
		let cancelled = false;
		ensureAnonymousUser()
			.then(() => {
				if (!cancelled) {
					setReady(true);
					setError(null);
				}
			})
			.catch((e: unknown) => {
				if (!cancelled) {
					setError(e instanceof Error ? e.message : "auth failed");
					setReady(false);
				}
			});
		return () => {
			cancelled = true;
		};
	}, []);

	return <AuthContext.Provider value={{ ready, error }}>{children}</AuthContext.Provider>;
}

export function useAuthReady() {
	return useContext(AuthContext);
}
