"use client";

import { useEffect, useRef } from "react";
import { registerUnityIframe, unregisterUnityIframe } from "@/lib/unityBridge";

type UnityLogMessage = {
	type: "unity-log";
	level: string;
	message: string;
};

type UnityReadyMessage = {
	type: "unity-ready";
};

type UnityProgressMessage = {
	type: "unity-progress";
	progress: number;
};

function isUnityLogMessage(data: unknown): data is UnityLogMessage {
	return (
		typeof data === "object" &&
		data !== null &&
		"type" in data &&
		(data as UnityLogMessage).type === "unity-log"
	);
}

function isUnityReadyMessage(data: unknown): data is UnityReadyMessage {
	return (
		typeof data === "object" &&
		data !== null &&
		"type" in data &&
		(data as UnityReadyMessage).type === "unity-ready"
	);
}

function isUnityProgressMessage(data: unknown): data is UnityProgressMessage {
	return (
		typeof data === "object" &&
		data !== null &&
		"type" in data &&
		(data as UnityProgressMessage).type === "unity-progress" &&
		typeof (data as UnityProgressMessage).progress === "number"
	);
}

type UnityWebGLFrameProps = {
	className?: string;
	onUnityReady?: () => void;
	onUnityProgress?: (progress: number) => void;
};

export function UnityWebGLFrame({
	className = "",
	onUnityReady,
	onUnityProgress,
}: UnityWebGLFrameProps) {
	const iframeRef = useRef<HTMLIFrameElement | null>(null);
	const onUnityReadyRef = useRef(onUnityReady);
	const onUnityProgressRef = useRef(onUnityProgress);
	onUnityReadyRef.current = onUnityReady;
	onUnityProgressRef.current = onUnityProgress;

	useEffect(() => {
		registerUnityIframe(iframeRef.current);
		let readyNotified = false;

		const notifyReady = () => {
			if (readyNotified) return;
			readyNotified = true;
			onUnityProgressRef.current?.(1);
			onUnityReadyRef.current?.();
		};

		const handleMessage = (event: MessageEvent) => {
			if (event.source !== iframeRef.current?.contentWindow) {
				return;
			}
			if (isUnityProgressMessage(event.data)) {
				const p = Math.max(0, Math.min(1, event.data.progress));
				onUnityProgressRef.current?.(p);
				return;
			}
			// Unity C# JsBridge がスプラッシュ後に送る
			if (isUnityReadyMessage(event.data)) {
				notifyReady();
				return;
			}
			if (!isUnityLogMessage(event.data)) return;
			const { level, message } = event.data;
			const prefix = "[Unity]";
			if (level === "error" || level === "exception") {
				console.error(prefix, message);
			} else if (level === "warning") {
				console.warn(prefix, message);
			} else {
				console.log(prefix, message);
			}
		};

		window.addEventListener("message", handleMessage);
		return () => {
			window.removeEventListener("message", handleMessage);
			unregisterUnityIframe(iframeRef.current);
		};
	}, []);

	return (
		<iframe
			ref={iframeRef}
			title="3D avatar view"
			data-aistalk-unity="embed"
			src="/unity/index.html"
			className={className}
			allow="autoplay; fullscreen"
		/>
	);
}
