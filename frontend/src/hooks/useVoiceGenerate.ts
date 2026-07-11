"use client";

import { useRef, useCallback } from "react";
import { useTTSLoading } from "@/contexts/TTSLoadingContext";
import { authFetch, withAccessToken } from "@/lib/apiClient";
import type { GenerateVoiceData } from "@/types/GenerateVoiceData";
import type { GenerateVoicePayload } from "@/types/GenerateVoicePayload";

const API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL || "";

export type TtsProvider = "aivis";

export interface UseVoiceGenerateOptions {
	onCommunicationError?: () => void;
	onQuotaExceeded?: () => void;
	/** SSE play イベント受信時に expression/motion / 話者 ID を Unity へ反映するコールバック */
	onPlayChunk?: (data: {
		expression?: string;
		motion?: string;
		speakerSpacePersonaId?: string;
		/** サーバー SSE の index（モーション API 結果と突き合わせ用） */
		segmentIndex?: number;
	}) => void;
}

declare global {
	interface Window {
		__TTSVolume?: number;
		__TTSPlaying?: boolean;
		onTTSPlaybackStart?: () => void;
		onTTSPlaybackEnd?: () => void;
	}
}

function getRMSVolume(dataArray: Uint8Array): number {
	let sumSq = 0;
	for (let i = 0; i < dataArray.length; i++) {
		const d = dataArray[i] - 128;
		sumSq += d * d;
	}
	const rms = Math.sqrt(sumSq / dataArray.length) / 128;
	return Math.min(1, rms * 2);
}

type AudioQueueItem = {
	audioData: string;
	expression?: string;
	motion?: string;
	speakerSpacePersonaId?: string;
	segmentIndex?: number;
};

export function useVoiceGenerate(options: UseVoiceGenerateOptions = {}) {
	const { onCommunicationError, onQuotaExceeded, onPlayChunk } = options;
	const onErrorRef = useRef(onCommunicationError);
	onErrorRef.current = onCommunicationError;
	const onQuotaRef = useRef(onQuotaExceeded);
	onQuotaRef.current = onQuotaExceeded;
	const onPlayChunkRef = useRef(onPlayChunk);
	onPlayChunkRef.current = onPlayChunk;

	const { setTTSLoading } = useTTSLoading();
	const sessionIdRef = useRef<string | null>(null);
	const eventSource = useRef<EventSource | null>(null);
	const audioQueue = useRef<AudioQueueItem[]>([]);
	const audioContext = useRef<AudioContext | null>(null);
	const isPlaying = useRef<boolean>(false);
	const hasCalledPlaybackStart = useRef<boolean>(false);
	const volumeLoopId = useRef<number>(0);
	const streamComplete = useRef<boolean>(false);
	/** 1 回の synthesize がサーバーから SSE complete を送り終えるまで待つ（複数エンジンを順に叩くときの順序用） */
	const waitServerRoundResolveRef = useRef<(() => void) | null>(null);
	/** generateVoice 呼び出し時に指定（SSE には話者が含まれないため再生キューへ引き渡す） */
	const playbackSpeakerRef = useRef<string | undefined>(undefined);
	/** generateVoice を直列化（同時 POST による SSE play の取り違え防止） */
	const synthQueueTailRef = useRef<Promise<void>>(Promise.resolve());
	/** モーション API などが解決するまで play でキューに積むだけにする */
	const playbackGateResolvedRef = useRef<boolean>(true);
	const startRecognitionRef = useRef<(() => void) | null>(null);
	const playNextAudioRef = useRef<((startRecognition: () => void) => Promise<void>) | null>(null);
	/** SSE に index が無い場合の連番 */
	const sseSegmentIndexFallbackRef = useRef<number>(0);

	const playNextAudio = async (startRecognition: () => void) => {
		if (isPlaying.current || audioQueue.current.length === 0) return;

		const AC =
			window.AudioContext || (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext;

		if (!audioContext.current || audioContext.current.state === "closed") {
			audioContext.current = new AC();
		}

		const actx = audioContext.current;
		if (!actx) {
			isPlaying.current = false;
			return;
		}
		await actx.resume();

		isPlaying.current = true;

		const item = audioQueue.current.shift();
		if (!item) {
			isPlaying.current = false;
			return;
		}

		const cb = onPlayChunkRef.current;
		if (
			cb &&
			(item.expression ||
				item.motion ||
				item.speakerSpacePersonaId ||
				item.segmentIndex !== undefined)
		) {
			cb({
				expression: item.expression,
				motion: item.motion,
				speakerSpacePersonaId: item.speakerSpacePersonaId,
				segmentIndex: item.segmentIndex,
			});
		}

		const binary = atob(item.audioData);
		const len = binary.length;
		const bytes = new Uint8Array(len);
		for (let i = 0; i < len; i++) {
			bytes[i] = binary.charCodeAt(i);
		}

		try {
			const audioBuffer = await actx.decodeAudioData(bytes.buffer.slice(0));
			const source = actx.createBufferSource();
			source.buffer = audioBuffer;

			const analyser = actx.createAnalyser();
			analyser.fftSize = 256;
			source.connect(analyser);
			analyser.connect(actx.destination);

			let sourceEnded = false;
			const dataArray = new Uint8Array(analyser.frequencyBinCount);

			const runVolumeLoop = () => {
				if (sourceEnded) {
					window.__TTSVolume = 0;
					volumeLoopId.current = 0;
					return;
				}
				analyser.getByteTimeDomainData(dataArray);
				window.__TTSVolume = getRMSVolume(dataArray);
				window.__TTSPlaying = true;
				volumeLoopId.current = requestAnimationFrame(runVolumeLoop);
			};

			source.onended = function () {
				sourceEnded = true;
				isPlaying.current = false;
				window.__TTSVolume = 0;
				window.__TTSPlaying = false;

				if (audioQueue.current.length === 0 && streamComplete.current) {
					hasCalledPlaybackStart.current = false;
					window.onTTSPlaybackEnd?.();
					setTTSLoading(false);
				}

				void playNextAudioRef.current?.(startRecognition);
			};

			if (!hasCalledPlaybackStart.current) {
				hasCalledPlaybackStart.current = true;
				window.onTTSPlaybackStart?.();
			}
			source.start(0);
			runVolumeLoop();
		} catch (e) {
			console.error("WebAudioデコードエラー:", e);
			isPlaying.current = false;
			window.__TTSVolume = 0;
			window.__TTSPlaying = false;
			void playNextAudioRef.current?.(startRecognition);
		}
	};

	playNextAudioRef.current = playNextAudio;

	const sessionStart = useCallback(
		async (startRecognition: () => void) => {
			startRecognitionRef.current = startRecognition;
			sessionIdRef.current = "sess_" + Date.now() + "_" + Math.floor(Math.random() * 10000);

			if (eventSource.current) {
				eventSource.current.close();
			}

			const sseUrl = await withAccessToken(
				API_BASE + `/api/event/stream/session/start?sessionId=${sessionIdRef.current}`,
			);
			eventSource.current = new EventSource(sseUrl);
			eventSource.current.onmessage = function () {};
			eventSource.current.addEventListener("play", function (event: MessageEvent) {
				try {
					const response = JSON.parse(event.data);
					if (response.audioData) {
						const segIdx =
							typeof response.index === "number"
								? response.index
								: sseSegmentIndexFallbackRef.current++;
						audioQueue.current.push({
							audioData: response.audioData,
							expression: response.expression,
							motion: response.motion,
							speakerSpacePersonaId: playbackSpeakerRef.current,
							segmentIndex: segIdx,
						});
						if (playbackGateResolvedRef.current) {
							const sr = startRecognitionRef.current;
							if (sr) void playNextAudioRef.current?.(sr);
						}
					}
				} catch (e) {
					console.error("SSEデータパースエラー:", e);
					eventSource.current?.close();
					eventSource.current = null;
					onErrorRef.current?.();
					setTTSLoading(false);
					const w = waitServerRoundResolveRef.current;
					waitServerRoundResolveRef.current = null;
					w?.();
				}
			});

			eventSource.current.addEventListener("complete", function (event: MessageEvent) {
				try {
					const response = JSON.parse(event.data);
					streamComplete.current = true;
					if (response.audioData) {
						const segIdx =
							typeof response.index === "number"
								? response.index
								: sseSegmentIndexFallbackRef.current++;
						audioQueue.current.push({
							audioData: response.audioData,
							expression: response.expression,
							motion: response.motion,
							speakerSpacePersonaId: playbackSpeakerRef.current,
							segmentIndex: segIdx,
						});
						if (playbackGateResolvedRef.current) {
							const sr = startRecognitionRef.current;
							if (sr) void playNextAudioRef.current?.(sr);
						}
					} else {
						if (playbackGateResolvedRef.current) {
							const sr = startRecognitionRef.current;
							if (sr) void playNextAudioRef.current?.(sr);
						}
						if (audioQueue.current.length === 0 && !isPlaying.current) {
							hasCalledPlaybackStart.current = false;
							window.onTTSPlaybackEnd?.();
							setTTSLoading(false);
						}
					}
				} catch (e) {
					console.error("SSEデータパースエラー:", e);
					eventSource.current?.close();
					eventSource.current = null;
					onErrorRef.current?.();
					setTTSLoading(false);
				} finally {
					const w = waitServerRoundResolveRef.current;
					waitServerRoundResolveRef.current = null;
					w?.();
				}
			});
			eventSource.current.onerror = function (e) {
				console.error("SSEエラー:", e);
				eventSource.current?.close();
				eventSource.current = null;
				onErrorRef.current?.();
				setTTSLoading(false);
				const w = waitServerRoundResolveRef.current;
				waitServerRoundResolveRef.current = null;
				w?.();
			};
		},
		[setTTSLoading],
	);

	const generateVoice = useCallback(
		async (
			generateVoiceData: GenerateVoiceData[],
			opts?: {
				provider?: TtsProvider;
				speakerSpacePersonaId?: string;
				voiceMasterId?: string;
				/** 解決するまで SSE の play でキューに積むだけ（口パク・再生はゲート解除後） */
				deferPlaybackUntil?: Promise<unknown>;
			},
		) => {
			const run = synthQueueTailRef.current.then(async () => {
				streamComplete.current = false;
				sseSegmentIndexFallbackRef.current = 0;
				const sp = (opts?.speakerSpacePersonaId ?? "").trim();
				playbackSpeakerRef.current = sp !== "" ? sp : undefined;
				setTTSLoading(true);

				const gate = opts?.deferPlaybackUntil;
				playbackGateResolvedRef.current = gate == null;
				if (gate) {
					void gate.finally(() => {
						playbackGateResolvedRef.current = true;
						const sr = startRecognitionRef.current;
						if (sr) void playNextAudioRef.current?.(sr);
					});
				}

				try {
					const currentSessionId = sessionIdRef.current;
					const vm = (opts?.voiceMasterId ?? "").trim();
					const updatedPayload: GenerateVoicePayload = {
						sessionId: currentSessionId || "",
						...(vm !== "" ? { voice_master_id: vm } : {}),
						segments: generateVoiceData,
					};

					const response = await authFetch(`${API_BASE}/api/event/tts/aivis/synthesize`, {
						method: "POST",
						headers: { "Content-Type": "application/json" },
						body: JSON.stringify(updatedPayload),
					});
					if (response.status === 429) {
						onQuotaRef.current?.();
						setTTSLoading(false);
						return;
					}
					const data = await response.json();
					if (data.status === "accepted") {
						console.log("音声合成リクエストが成功しました");
						await new Promise<void>((resolve) => {
							waitServerRoundResolveRef.current = resolve;
						});
					} else {
						console.error("音声合成リクエストに失敗しました");
						onErrorRef.current?.();
						setTTSLoading(false);
					}
				} catch (error) {
					console.error("音声合成エラー:", error);
					const w = waitServerRoundResolveRef.current;
					waitServerRoundResolveRef.current = null;
					w?.();
					onErrorRef.current?.();
					setTTSLoading(false);
				}
			});
			synthQueueTailRef.current = run.then(() => {}).catch(() => {});
			await run;
		},
		[setTTSLoading],
	);

	return {
		generateVoice,
		sessionStart,
	};
}
