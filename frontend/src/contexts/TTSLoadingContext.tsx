"use client";

import React, { createContext, useContext, useState } from "react";

const TTSLoadingContext = createContext<{
	ttsLoading: boolean;
	setTTSLoading: React.Dispatch<React.SetStateAction<boolean>>;
} | null>(null);

export const TTSLoadingProvider = ({ children }: { children: React.ReactNode }) => {
	const [ttsLoading, setTTSLoading] = useState(false);

	return (
		<TTSLoadingContext.Provider value={{ ttsLoading, setTTSLoading }}>
			{children}
		</TTSLoadingContext.Provider>
	);
};

export const useTTSLoading = () => {
	const context = useContext(TTSLoadingContext);
	if (!context) {
		throw new Error("useTTSLoading must be used within a TTSLoadingProvider");
	}
	return context;
};
