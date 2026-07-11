"use client";

import { Suspense, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useChatGPT } from "@/hooks/useChatGPT";
import { useVoiceRecognition } from "@/hooks/useVoiceRecognition";
import { useMicVolumeAnalyzer } from "@/hooks/useMicVolumeAnalyzer";
import { useVoiceGenerate } from "@/hooks/useVoiceGenerate";
import { CallOpenAIError } from "@/exceptions/CallOpenAIError";
import { useTTSLoading } from "@/contexts/TTSLoadingContext";
import { useAuthReady } from "@/contexts/AuthContext";
import type { Message } from "@/types/Message";
import type { FetchOpenAIResponse } from "@/types/FetchOpenAIResponse";
import type { GenerateVoiceData } from "@/types/GenerateVoiceData";
import { MicVolumeRing } from "@/components/MicVolumeRing";
import { SelfViewBackgroundPalette } from "@/components/SelfViewBackgroundPalette";
import { UnityWebGLFrame } from "@/components/UnityWebGLFrame";
import { Modal } from "@/components/Modal";
import {
	buildSetMouthOpenCommandJson,
	sendJsonToUnityReceiveController,
	sendFacialParametersJsonToUnity,
	sendFocusSpacePersonaToUnity,
	sendMotionExpressionToUnity,
	sendPlayMotionToUnity,
	sendSelfViewBackgroundToUnity,
	sendSpacePersonasToUnity,
} from "@/lib/unityBridge";
import {
	resolveBackgroundPreset,
	saveBackgroundPresetIdToStorage,
	type SelfViewBackgroundPreset,
} from "@/lib/selfViewBackgroundPresets";
import { KISS_MOTION_FACE_PRESET } from "@/lib/motionExpressionFacialPresets";
import { inferAvatarMotion } from "@/api/avatarMotionApi";
import { getDemoPersona, getDemoSelfAvatar, MESSAGES_LS_KEY } from "@/lib/demoConfig";
import {
	applyPersonaHoverOutlineVisible,
	loadPersonaHoverOutlineVisible,
} from "@/lib/personaHoverOutline";
import { resolveModelUrl } from "@/lib/resolveModelUrl";
import { stringifyChatTurnContent, formatChatMessageForDisplay } from "@/utils/chatTurnContent";

const DELEGATE_LINE_SEGMENTATION_TO_MOTION_AI = true;
const MAX_HISTORY_MESSAGES = 40;

function loadMessagesFromStorage(): Message[] {
	if (typeof window === "undefined") return [];
	try {
		const raw = localStorage.getItem(MESSAGES_LS_KEY);
		if (!raw) return [];
		const parsed = JSON.parse(raw) as Message[];
		return Array.isArray(parsed) ? parsed : [];
	} catch {
		return [];
	}
}

function saveMessagesToStorage(messages: Message[]) {
	if (typeof window === "undefined") return;
	try {
		localStorage.setItem(MESSAGES_LS_KEY, JSON.stringify(messages));
	} catch {
		// ignore quota errors
	}
}

function trimMessagesForRequest(messages: Message[]): Message[] {
	if (messages.length <= MAX_HISTORY_MESSAGES) return messages;
	return messages.slice(-MAX_HISTORY_MESSAGES);
}

function ChatPageInner() {
	const selfAvatar = useMemo(() => getDemoSelfAvatar(), []);
	const persona = useMemo(() => getDemoPersona(), []);
	const { ttsLoading, setTTSLoading } = useTTSLoading();
	const { ready: authReady, error: authError } = useAuthReady();
	const { sendMessage } = useChatGPT();
	const [messages, setMessages] = useState<Message[]>([]);
	const messagesRef = useRef<Message[]>(messages);
	messagesRef.current = messages;
	const [inputValue, setInputValue] = useState("");
	const [unityReady, setUnityReady] = useState(false);
	const [leftDrawerOpen, setLeftDrawerOpen] = useState(false);
	const [rightDrawerOpen, setRightDrawerOpen] = useState(false);
	const [backgroundPreset, setBackgroundPreset] = useState<SelfViewBackgroundPreset>(() =>
		resolveBackgroundPreset(),
	);
	const [hoverOutlineVisible, setHoverOutlineVisible] = useState(() => loadPersonaHoverOutlineVisible());
	const backgroundPresetRef = useRef(backgroundPreset);
	backgroundPresetRef.current = backgroundPreset;
	const [errorOpen, setErrorOpen] = useState(false);
	const [errorKind, setErrorKind] = useState<"communication" | "quota" | "mic">("communication");
	const resumeVoiceAfterTtsRef = useRef(false);
	const motionLinesForPlaybackRef = useRef<{ motion: string; expression: string }[]>([]);
	const mouthVolumeSmoothedRef = useRef(0);
	const lastSentMouthOpenRef = useRef(-1);

	const sendUnityMouthOpen = (value: number) => {
		const normalized = Math.max(0, Math.min(1, value));
		if (Math.abs(lastSentMouthOpenRef.current - normalized) < 0.01) return;
		lastSentMouthOpenRef.current = normalized;
		sendJsonToUnityReceiveController(buildSetMouthOpenCommandJson(normalized));
	};

	const showCommunicationError = () => {
		sendUnityMouthOpen(0);
		sendPlayMotionToUnity("idle", true);
		setTTSLoading(false);
		setErrorKind("communication");
		setErrorOpen(true);
	};

	const showQuotaExceededError = () => {
		sendUnityMouthOpen(0);
		sendPlayMotionToUnity("idle", true);
		setTTSLoading(false);
		setErrorKind("quota");
		setErrorOpen(true);
	};

	const { generateVoice, sessionStart } = useVoiceGenerate({
		onCommunicationError: showCommunicationError,
		onQuotaExceeded: showQuotaExceededError,
		onPlayChunk: (data) => {
			const sid = data.speakerSpacePersonaId?.trim() || persona.id;
			sendFocusSpacePersonaToUnity(sid);
			const idx = typeof data.segmentIndex === "number" ? data.segmentIndex : 0;
			const ov = motionLinesForPlaybackRef.current[idx];
			const expression = ov?.expression ?? data.expression;
			const motion = ov?.motion ?? data.motion;
			if ((motion ?? "").trim().toLowerCase() === "kiss") {
				sendFacialParametersJsonToUnity(KISS_MOTION_FACE_PRESET, sid);
			} else if (expression) {
				sendMotionExpressionToUnity(expression, sid);
			}
			if (motion) sendPlayMotionToUnity(motion, true, sid, idx);
		},
	});

	const handleSendRef = useRef<((text?: string) => Promise<void>) | null>(null);
	const { startRecognition, stopRecognition, isRecognizing, transcript, clearTranscript } =
		useVoiceRecognition({
			onSilenceCommit: (text) => handleSendRef.current?.(text),
			onError: (reason) => {
				if (reason === "not-allowed") {
					setErrorKind("mic");
					setErrorOpen(true);
				} else if (reason === "network") {
					console.warn(
						"[voiceRecognition] Chrome Web Speech の network エラー。マイク権限・ネット・ブラウザを確認してください。",
					);
					setErrorKind("mic");
					setErrorOpen(true);
				}
			},
		});
	const { startMicVolume, stopMicVolume, volume: micVolumeLevel } = useMicVolumeAnalyzer();

	const applyBackgroundPreset = useCallback((preset: SelfViewBackgroundPreset) => {
		setBackgroundPreset(preset);
		saveBackgroundPresetIdToStorage(preset.id);
		sendSelfViewBackgroundToUnity(preset.payload);
	}, []);

	const toggleHoverOutlineVisible = useCallback((visible: boolean) => {
		setHoverOutlineVisible(visible);
		applyPersonaHoverOutlineVisible(visible);
	}, []);

	useEffect(() => {
		setMessages(loadMessagesFromStorage());
	}, []);

	useEffect(() => {
		saveMessagesToStorage(messages);
	}, [messages]);

	useEffect(() => {
		setInputValue(transcript);
	}, [transcript]);

	const unityPersonasSyncedRef = useRef(false);
	const unityBackgroundSyncedRef = useRef(false);
	const unityHoverOutlineSyncedRef = useRef(false);

	useEffect(() => {
		if (!unityReady || unityPersonasSyncedRef.current) return;
		unityPersonasSyncedRef.current = true;
		sendSpacePersonasToUnity({
			spaceId: persona.spaceId,
			focusSpacePersonaId: persona.id,
			personas: [
				{
					spacePersonaId: selfAvatar.id,
					personaId: selfAvatar.personaId,
					modelUrl: resolveModelUrl(selfAvatar.modelUrl),
					label: selfAvatar.label,
					personaKind: "self",
				},
				{
					spacePersonaId: persona.id,
					personaId: persona.personaId,
					modelUrl: resolveModelUrl(persona.modelUrl),
					label: persona.nickname,
					personaKind: "owned",
				},
			],
		});
		sendFocusSpacePersonaToUnity(persona.id);
	}, [unityReady, persona, selfAvatar]);

	useEffect(() => {
		if (!unityReady || unityBackgroundSyncedRef.current) return;
		unityBackgroundSyncedRef.current = true;
		sendSelfViewBackgroundToUnity(backgroundPresetRef.current.payload);
	}, [unityReady]);

	useEffect(() => {
		if (!unityReady || unityHoverOutlineSyncedRef.current) return;
		unityHoverOutlineSyncedRef.current = true;
		applyPersonaHoverOutlineVisible(hoverOutlineVisible);
	}, [unityReady, hoverOutlineVisible]);

	useEffect(() => {
		if (isRecognizing) {
			void startMicVolume().catch(() => {});
		} else {
			stopMicVolume();
		}
	}, [isRecognizing, startMicVolume, stopMicVolume]);

	useEffect(() => {
		if (!authReady) return;
		void sessionStart(startRecognition);
	}, [authReady, sessionStart, startRecognition]);

	useEffect(() => {
		if (!ttsLoading) {
			sendPlayMotionToUnity("idle", true);
			sendUnityMouthOpen(0);
			if (resumeVoiceAfterTtsRef.current) startRecognition();
			return;
		}
		sendPlayMotionToUnity("thinking", true);
	}, [ttsLoading, startRecognition]);

	useEffect(() => {
		if (!ttsLoading) return;
		let raf = 0;
		const tick = () => {
			const vol = typeof window !== "undefined" ? window.__TTSVolume ?? 0 : 0;
			mouthVolumeSmoothedRef.current += (vol - mouthVolumeSmoothedRef.current) * 0.35;
			sendUnityMouthOpen(mouthVolumeSmoothedRef.current);
			raf = requestAnimationFrame(tick);
		};
		raf = requestAnimationFrame(tick);
		return () => cancelAnimationFrame(raf);
	}, [ttsLoading]);

	const playResponse = useCallback(
		async (response: FetchOpenAIResponse, userMessage: string) => {
			const script = response.script;
			if (!script?.lines?.length) return;

			const styleRoot = script.style_local_id?.trim() ?? "";
			const motionOpts = {
				...(styleRoot ? { style_local_id: styleRoot } : {}),
				...(userMessage.trim() ? { last_user_message: userMessage.trim() } : {}),
			};
			const fullReply = script.lines.map((l) => l.text).join("\n").trim();
			if (!fullReply) return;

			if (DELEGATE_LINE_SEGMENTATION_TO_MOTION_AI) {
				motionLinesForPlaybackRef.current = [];
				const motionRes = await inferAvatarMotion({
					lines: [{ text: fullReply }],
					delegate_segmentation: true,
					...motionOpts,
				});
				const pairs = motionRes.lines
					.map((ln) => ({
						motion: ln.motion,
						expression: ln.expression,
						message: (ln.text ?? "").trim(),
					}))
					.filter((p) => p.message.length > 0);
				if (pairs.length === 0) return;
				motionLinesForPlaybackRef.current = pairs.map(({ motion, expression }) => ({ motion, expression }));
				const segments: GenerateVoiceData[] = pairs.map((p) => ({
					message: p.message,
					...(styleRoot ? { style_local_id: styleRoot } : {}),
				}));
				await generateVoice(segments, {
					provider: "aivis",
					speakerSpacePersonaId: persona.id,
					...(persona.voiceMasterId ? { voiceMasterId: persona.voiceMasterId } : {}),
				});
			}
		},
		[generateVoice, persona.id, persona.voiceMasterId],
	);

	const handleSend = useCallback(
		async (overrideMessage?: string) => {
			const hadMic = isRecognizing;
			resumeVoiceAfterTtsRef.current = overrideMessage !== undefined || hadMic;
			const rawMessage = (overrideMessage ?? inputValue).trim();
			if (!rawMessage) return;

			setInputValue("");
			stopRecognition();
			clearTranscript();
			sessionStart(startRecognition);

			const createdAt = new Date().toISOString();
			const userContent = stringifyChatTurnContent({
				id: selfAvatar.id,
				name: selfAvatar.label,
				message: rawMessage,
			});
			const prior = trimMessagesForRequest(messagesRef.current);
			const requestMessages: Message[] = [...prior, { role: "user", content: userContent, createdAt }];
			const snapshot = messagesRef.current.slice();
			setMessages(requestMessages);

			try {
				const response = await sendMessage(requestMessages, {
					space_persona_id: persona.id,
					participants: [
						{
							space_persona_id: selfAvatar.id,
							label: selfAvatar.label,
							persona_kind: "self",
						},
					],
					persona: {
						name: persona.personaName,
						personality: persona.personality,
						response_style: persona.responseStyle,
					},
				});
				setMessages(response.histories);
				await playResponse(response, rawMessage);
			} catch (e) {
				console.error(e);
				setMessages(snapshot);
				if (e instanceof CallOpenAIError && e.isQuotaExceeded) {
					showQuotaExceededError();
				} else {
					showCommunicationError();
				}
			}
		},
		[
			clearTranscript,
			inputValue,
			isRecognizing,
			persona,
			selfAvatar,
			playResponse,
			sendMessage,
			sessionStart,
			startRecognition,
			stopRecognition,
		],
	);

	useEffect(() => {
		handleSendRef.current = handleSend;
	}, [handleSend]);

	return (
		<div className="relative h-screen w-screen overflow-hidden bg-black">
			{authError ? (
				<div className="absolute left-0 right-0 top-0 z-[60] bg-red-700/90 px-4 py-2 text-center text-sm text-white">
					認証に失敗しました: {authError}
				</div>
			) : null}
			<div className="absolute inset-0 z-0">
				<UnityWebGLFrame
					className="h-full w-full border-0 bg-black"
					onUnityReady={() => setUnityReady(true)}
				/>
			</div>

			{leftDrawerOpen || rightDrawerOpen ? (
				<button
					type="button"
					aria-label="メニューを閉じる"
					className="fixed inset-0 z-40 bg-black/25"
					onClick={() => {
						setLeftDrawerOpen(false);
						setRightDrawerOpen(false);
					}}
				/>
			) : null}

			<button
				type="button"
				onClick={() => {
					setRightDrawerOpen(false);
					setLeftDrawerOpen((o) => !o);
				}}
				className="fixed left-0 top-1/2 z-20 flex h-24 w-10 -translate-y-1/2 items-center justify-center rounded-r-2xl bg-black/70 text-white backdrop-blur-sm hover:bg-black/90"
				title="背景"
			>
				<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
					<circle cx="12" cy="12" r="9" />
					<path d="M12 3v18" />
					<path d="M3 12h18" />
				</svg>
			</button>

			<div
				className={`fixed left-0 top-0 bottom-0 z-50 flex w-[min(280px,82vw)] flex-col bg-black/90 backdrop-blur-md transition-transform duration-300 ${leftDrawerOpen ? "translate-x-0" : "-translate-x-full"}`}
			>
				<div className="border-b border-white/10 px-4 py-3">
					<p className="text-sm font-semibold text-white">表示</p>
					<p className="mt-0.5 text-xs text-white/50">背景とアバター枠線</p>
				</div>
				<div className="flex-1 overflow-y-auto p-4 space-y-6">
					<section>
						<p className="mb-2 text-xs font-semibold tracking-wide text-white/55">背景</p>
						<SelfViewBackgroundPalette
							selectedId={backgroundPreset.id}
							onSelect={applyBackgroundPreset}
						/>
					</section>
					<section>
						<p className="mb-2 text-xs font-semibold tracking-wide text-white/55">アバター枠線</p>
						<p className="mb-3 text-xs text-white/45">アバターを選択・ホバーしたときの光る枠</p>
						<button
							type="button"
							role="switch"
							aria-checked={hoverOutlineVisible}
							onClick={() => toggleHoverOutlineVisible(!hoverOutlineVisible)}
							className={`flex w-full items-center justify-between rounded-xl border px-3 py-3 text-sm transition-colors ${
								hoverOutlineVisible
									? "border-pink-300/40 bg-pink-300/10 text-white"
									: "border-white/10 bg-white/5 text-white/80"
							}`}
						>
							<span>選択枠を表示</span>
							<span
								className={`rounded-full px-2 py-0.5 text-xs ${
									hoverOutlineVisible ? "bg-pink-300/20 text-pink-100" : "bg-white/10 text-white/50"
								}`}
							>
								{hoverOutlineVisible ? "ON" : "OFF"}
							</span>
						</button>
					</section>
				</div>
			</div>

			<button
				type="button"
				onClick={() => {
					setLeftDrawerOpen(false);
					setRightDrawerOpen((o) => !o);
				}}
				className="fixed right-0 top-1/2 z-20 flex h-24 w-10 -translate-y-1/2 items-center justify-center rounded-l-2xl bg-black/70 text-white backdrop-blur-sm hover:bg-black/90"
				title="会話履歴"
			>
				<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
					<path d="M15 18l-6-6 6-6" />
				</svg>
			</button>

			<div
				className={`fixed right-0 top-0 bottom-0 z-50 flex w-[min(320px,88vw)] flex-col bg-black/90 backdrop-blur-md transition-transform duration-300 ${rightDrawerOpen ? "translate-x-0" : "translate-x-full"}`}
			>
				<div className="flex items-center justify-between border-b border-white/10 px-4 py-3">
					<p className="text-sm font-semibold text-white">会話履歴</p>
					<button
						type="button"
						className="text-xs text-white/60 hover:text-white"
						onClick={() => {
							setMessages([]);
							localStorage.removeItem(MESSAGES_LS_KEY);
						}}
					>
						クリア
					</button>
				</div>
				<div className="flex-1 overflow-y-auto p-3 space-y-2">
					{messages.length === 0 ? (
						<p className="text-sm text-white/50">まだ会話がありません</p>
					) : (
						messages.map((m, i) => {
							const isUser = m.role === "user";
							const { name, message } = formatChatMessageForDisplay(
								m.content,
								isUser ? "あなた" : persona.nickname,
							);
							return (
								<div
									key={`${m.createdAt ?? i}-${i}`}
									className={`rounded-lg p-3 text-sm ${isUser ? "ml-4 bg-white/20" : "mr-4 bg-white/10"}`}
								>
									<span className="mb-1 block text-xs font-semibold text-white/70">
										{name}
									</span>
									<p className="whitespace-pre-line text-white/95">{message}</p>
								</div>
							);
						})
					)}
				</div>
			</div>

			<div className="fixed bottom-6 left-1/2 z-40 flex -translate-x-1/2 flex-col items-center gap-2">
				<div className="flex items-center gap-3">
					<div className="relative flex h-[62px] w-[62px] items-center justify-center">
						{isRecognizing ? <MicVolumeRing volume={micVolumeLevel} /> : null}
						<button
							type="button"
							aria-pressed={isRecognizing}
							onClick={() => (isRecognizing ? stopRecognition() : startRecognition())}
							className={`relative z-10 flex h-12 w-12 items-center justify-center rounded-full border-2 shadow-lg backdrop-blur-sm ${
								isRecognizing
									? "border-white bg-red-600 text-white"
									: "border-white/30 bg-black/45 text-white/70"
							}`}
						>
							<svg xmlns="http://www.w3.org/2000/svg" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden>
								<path d="M12 14a3 3 0 0 0 3-3V5a3 3 0 0 0-6 0v6a3 3 0 0 0 3 3z" />
								<path d="M19 10v1a7 7 0 0 1-14 0v-1M12 18v4M8 22h8" />
							</svg>
						</button>
					</div>
				</div>
				<div className="w-[min(480px,92vw)] rounded-[1.75rem] border border-neutral-900/45 bg-neutral-600/92 px-3 py-2.5 shadow-lg backdrop-blur-sm">
					<div className="flex items-center gap-2">
						<input
							type="text"
							value={inputValue}
							onChange={(e) => setInputValue(e.target.value)}
							onKeyDown={(e) => {
								if (e.key === "Enter" && !e.nativeEvent.isComposing) void handleSend();
							}}
							placeholder={`${persona.personaName}に話しかける…`}
							className="min-w-0 flex-1 rounded-2xl bg-transparent px-3 py-2.5 text-sm text-white placeholder:text-white/40 focus:outline-none"
						/>
						<button
							type="button"
							aria-label="送信"
							disabled={ttsLoading || !inputValue.trim()}
							onClick={() => void handleSend()}
							className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-white text-black shadow-md transition-opacity disabled:opacity-40"
						>
							<svg
								xmlns="http://www.w3.org/2000/svg"
								width="18"
								height="18"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								strokeWidth="2"
								strokeLinecap="round"
								strokeLinejoin="round"
								aria-hidden
							>
								<path d="M22 2 11 13" />
								<path d="m22 2-7 20-4-9-9-4 20-7z" />
							</svg>
						</button>
					</div>
				</div>
			</div>

			<Modal
				isOpen={errorOpen}
				onClose={() => setErrorOpen(false)}
				title={
					errorKind === "quota" ? "利用上限" : errorKind === "mic" ? "マイク" : "通信エラー"
				}
			>
				{errorKind === "quota" ? (
					<p className="text-gray-700">
						デモ版の上限に達しました。
						<br />
						一日置いてから遊びにきてね。
					</p>
				) : errorKind === "mic" ? (
					<p className="text-gray-700">
						マイクを使えませんでした。
						<br />
						ブラウザのサイト設定でマイクを許可し、Chrome で再度お試しください。
					</p>
				) : (
					<p className="text-gray-700">通信に失敗しました。時間を空けてもう一度お試しください。</p>
				)}
			</Modal>
		</div>
	);
}

export default function ChatPageContent() {
	return (
		<Suspense fallback={<div className="h-screen w-screen bg-black" aria-hidden />}>
			<ChatPageInner />
		</Suspense>
	);
}
