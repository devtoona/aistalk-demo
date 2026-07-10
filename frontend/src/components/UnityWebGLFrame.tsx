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

    const handleMessage = (event: MessageEvent) => {
      if (event.source !== iframeRef.current?.contentWindow) {
        return;
      }
      if (isUnityReadyMessage(event.data)) {
        onUnityReadyRef.current?.();
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
