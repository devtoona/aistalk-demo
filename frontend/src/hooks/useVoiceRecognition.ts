"use client";

import { useState, useRef, useEffect, useCallback } from 'react';

export type VoiceRecognitionResult = {
	startRecognition: () => void;
	stopRecognition: () => void;
	isRecognizing: boolean;
	transcript: string;
	clearTranscript: () => void;
};

interface SpeechRecognitionInstance {
	start: () => void;
	stop: () => void;
	continuous: boolean;
	interimResults: boolean;
	maxAlternatives: number;
	lang: string;
	onstart: (() => void) | null;
	onresult: ((event: { results: { length: number;[i: number]: { length: number;[j: number]: { transcript: string }; isFinal: boolean } }; }) => void) | null;
	onerror: ((event: { error?: string; message?: string }) => void) | null;
	onend: (() => void) | null;
}

export type UseVoiceRecognitionOptions = {
	onSilenceCommit?: (text: string) => void;
	onError?: (reason: string) => void;
};

const SILENCE_MS = 2000;

export function useVoiceRecognition(
	options: UseVoiceRecognitionOptions = {},
): VoiceRecognitionResult {
	const onSilenceCommitRef = useRef(options.onSilenceCommit);
	onSilenceCommitRef.current = options.onSilenceCommit;
	const onErrorRef = useRef(options.onError);
	onErrorRef.current = options.onError;

	const [isRecognizing, setIsRecognizing] = useState(false);
	const [transcript, setTranscript] = useState('');
	const transcriptRef = useRef('');
	const recognitionObjectRef = useRef<SpeechRecognitionInstance | null>(null);
	const silenceTimerRef = useRef<NodeJS.Timeout | null>(null);
	const isRecognizingRef = useRef<boolean>(false);

	const clearSilenceTimer = () => {
		if (silenceTimerRef.current) {
			clearTimeout(silenceTimerRef.current);
			silenceTimerRef.current = null;
		}
	};

	const resetTranscript = () => {
		transcriptRef.current = '';
		setTranscript('');
	};

	const stopRecognition = useCallback(() => {
		clearSilenceTimer();
		isRecognizingRef.current = false;
		setIsRecognizing(false);
		const rec = recognitionObjectRef.current;
		if (rec) {
			try {
				rec.stop();
			} catch {
				// ignore
			}
		}
		resetTranscript();
	}, []);

	const startRecognition = useCallback(() => {
		if (typeof window === 'undefined') return;
		const SR =
			(window as unknown as { SpeechRecognition?: new () => SpeechRecognitionInstance }).SpeechRecognition ||
			(window as unknown as { webkitSpeechRecognition?: new () => SpeechRecognitionInstance }).webkitSpeechRecognition;
		if (!SR) {
			console.warn('Web Speech API is not supported');
			return;
		}

		clearSilenceTimer();
		resetTranscript();

		const recognition = new SR();
		recognition.continuous = true;
		recognition.interimResults = true;
		recognition.maxAlternatives = 1;
		recognition.lang = 'ja-JP';

		recognition.onstart = () => {
			isRecognizingRef.current = true;
			setIsRecognizing(true);
		};

		recognition.onresult = (event) => {
			let interim = '';
			let finalText = '';
			for (let i = 0; i < event.results.length; i++) {
				const result = event.results[i];
				const text = result[0]?.transcript ?? '';
				if (result.isFinal) finalText += text;
				else interim += text;
			}
			const combined = (finalText + interim).trim();
			transcriptRef.current = combined;
			setTranscript(combined);

			clearSilenceTimer();
			if (combined) {
				silenceTimerRef.current = setTimeout(() => {
					const t = transcriptRef.current.trim();
					if (t && isRecognizingRef.current) {
						onSilenceCommitRef.current?.(t);
						stopRecognition();
					}
				}, SILENCE_MS);
			}
		};

		recognition.onerror = (event) => {
			const reason = event?.error || "unknown";
			console.warn("[voiceRecognition] error:", reason, event?.message ?? "");
			// aborted は stop() 時の想定動作
			if (reason !== "aborted") {
				onErrorRef.current?.(reason);
			}
			stopRecognition();
		};

		recognition.onend = () => {
			if (isRecognizingRef.current) {
				try {
					recognition.start();
				} catch {
					isRecognizingRef.current = false;
					setIsRecognizing(false);
				}
			}
		};

		recognitionObjectRef.current = recognition;
		try {
			recognition.start();
		} catch {
			stopRecognition();
		}
	}, [stopRecognition]);

	useEffect(() => {
		return () => {
			clearSilenceTimer();
			const rec = recognitionObjectRef.current;
			if (rec) {
				try {
					rec.stop();
				} catch {
					// ignore
				}
			}
		};
	}, []);

	const clearTranscript = useCallback(() => {
		resetTranscript();
	}, []);

	return {
		startRecognition,
		stopRecognition,
		isRecognizing,
		transcript,
		clearTranscript,
	};
}
