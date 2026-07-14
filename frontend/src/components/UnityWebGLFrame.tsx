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

type UnityWebGLFrameProps = {
	className?: string;
	onUnityReady?: () => void;
};

export function UnityWebGLFrame({
	className = "",
	onUnityReady,
}: UnityWebGLFrameProps) {
  const iframeRef = useRef<HTMLIFrameElement | null>(null);
  const onUnityReadyRef = useRef(onUnityReady);
  onUnityReadyRef.current = onUnityReady;

  useEffect(() => {
    registerUnityIframe(iframeRef.current);
    let readyNotified = false;
    let pollId = 0;

    const notifyReady = () => {
      if (readyNotified) return;
      readyNotified = true;
      if (pollId) window.clearInterval(pollId);
      onUnityReadyRef.current?.();
    };

    const handleMessage = (event: MessageEvent) => {
      if (event.source !== iframeRef.current?.contentWindow) {
        return;
      }
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

    // GCS 上の index.html が postMessage しない場合のフォールバック
    pollId = window.setInterval(() => {
      const win = iframeRef.current?.contentWindow as
        | (Window & { unityInstance?: unknown })
        | null;
      if (win?.unityInstance) {
        notifyReady();
      }
    }, 200);

    window.addEventListener("message", handleMessage);
    return () => {
      window.clearInterval(pollId);
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
