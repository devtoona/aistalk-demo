"use client";

import {
  facialPresetForMotionExpression,
  randomAmaeFacialPreset,
} from "./motionExpressionFacialPresets";
import { resolvePlayMotionName } from "./motionSemanticExpand";

type UnityInstance = {
  SendMessage: (gameObject: string, methodName: string, parameter?: string) => void;
};

type UnityFrameWindow = Window & {
  unityInstance?: UnityInstance;
};

let unityIframe: HTMLIFrameElement | null = null;

/** ChatScene / デモ UI の選択枠（薄いピンク・極細） */
export const CHAT_SCENE_HOVER_OUTLINE_COLOR = "#ffc8d6";
export const CHAT_SCENE_HOVER_OUTLINE_WIDTH = 0.004;

export type PersonaHoverOutlineStyle = {
  colorHtml: string;
  width: number;
};

export function registerUnityIframe(iframe: HTMLIFrameElement | null) {
  unityIframe = iframe;
}

export function unregisterUnityIframe(iframe: HTMLIFrameElement | null) {
  if (unityIframe === iframe) {
    unityIframe = null;
  }
}

export function buildSetMouthOpenCommandJson(value: number): string {
  return JSON.stringify({
    action: "setMouthOpen",
    payloadJson: JSON.stringify({
      value: Math.max(0, Math.min(1, value)),
    }),
  });
}

function appendSpacePersonaId(
  payload: Record<string, unknown>,
  spacePersonaId?: string | null
): Record<string, unknown> {
  if (spacePersonaId) {
    return { ...payload, spacePersonaId };
  }
  return payload;
}

export function buildSetExpressionCommandJson(
  name: string,
  value = 1,
  spacePersonaId?: string | null
): string {
  const body = appendSpacePersonaId(
    {
      name,
      value: Math.max(0, Math.min(1, value)),
    },
    spacePersonaId
  );
  return JSON.stringify({
    action: "setExpression",
    payloadJson: JSON.stringify(body),
  });
}

export type FacialPresetJson = {
  situation?: string;
  parameters: Array<{ Name: string; Value: number }>;
  blush?: number;
  eyes_dark?: number;
};

export type FacialPresetJsonForUnity = {
  situation?: string;
  parameters: Array<{ Name: string; Value: number }>;
};

export function buildSetFacialParametersJsonCommandJson(
  facial: FacialPresetJsonForUnity,
  spacePersonaId?: string | null
): string {
  const forUnity: FacialPresetJsonForUnity = {
    situation: facial.situation,
    parameters: facial.parameters ?? [],
  };
  const facialJson = JSON.stringify(forUnity);
  const body = appendSpacePersonaId({ facialJson }, spacePersonaId);
  return JSON.stringify({
    action: "setFacialParametersJson",
    payloadJson: JSON.stringify(body),
  });
}

export function sendFacialParametersJsonToUnity(
  facial: FacialPresetJson,
  spacePersonaId?: string | null
): boolean {
  const blush =
    typeof facial.blush === "number" ? Math.max(0, Math.min(1, facial.blush)) : 0;
  const eyes =
    typeof facial.eyes_dark === "number" ? Math.max(0, Math.min(1, facial.eyes_dark)) : 0;
  const ok = sendJsonToUnityReceiveController(
    buildSetFacialParametersJsonCommandJson(
      { situation: facial.situation, parameters: facial.parameters ?? [] },
      spacePersonaId
    )
  );
  sendBlushToUnity(blush, spacePersonaId);
  sendEyesDarkToUnity(eyes, spacePersonaId);
  return ok;
}

export function buildSetBlushCommandJson(value = 1, spacePersonaId?: string | null): string {
  const body = appendSpacePersonaId(
    { value: Math.max(0, Math.min(1, value)) },
    spacePersonaId
  );
  return JSON.stringify({
    action: "setBlush",
    payloadJson: JSON.stringify(body),
  });
}

export function buildSetEyesDarkCommandJson(value = 1, spacePersonaId?: string | null): string {
  const body = appendSpacePersonaId(
    { value: Math.max(0, Math.min(1, value)) },
    spacePersonaId
  );
  return JSON.stringify({
    action: "setEyesDark",
    payloadJson: JSON.stringify(body),
  });
}

export function buildPlayMotionCommandJson(
  name: string,
  loop = true,
  spacePersonaId?: string | null
): string {
  const body = appendSpacePersonaId({ name, loop }, spacePersonaId);
  return JSON.stringify({
    action: "playMotion",
    payloadJson: JSON.stringify(body),
  });
}

export function buildSetHoverOutlineStyleCommandJson(colorHtml: string, width: number): string {
  return JSON.stringify({
    action: "setHoverOutlineStyle",
    payloadJson: JSON.stringify({
      colorHtml,
      width: Math.max(0, width),
    }),
  });
}

export type UnityRgbaPayload = { r: number; g: number; b: number; a: number };

export type UnitySelfViewBackgroundPayload = {
  useProceduralSkybox: boolean;
  useGradient: boolean;
  main: UnityRgbaPayload;
  skyboxColorTop: UnityRgbaPayload;
  skyboxColorBottom: UnityRgbaPayload;
  usePanorama: boolean;
  panoramaUrl: string;
  panoramaFlipY: boolean;
};

export const DEFAULT_SELF_VIEW_BACKGROUND: UnitySelfViewBackgroundPayload = {
  useProceduralSkybox: true,
  useGradient: true,
  main: { r: 27 / 255, g: 42 / 255, b: 74 / 255, a: 1 },
  skyboxColorTop: { r: 27 / 255, g: 42 / 255, b: 74 / 255, a: 1 },
  skyboxColorBottom: { r: 6 / 255, g: 11 / 255, b: 20 / 255, a: 1 },
  usePanorama: false,
  panoramaUrl: "",
  panoramaFlipY: false,
};

export function buildSetSelfViewBackgroundCommandJson(
  payload: UnitySelfViewBackgroundPayload
): string {
  return JSON.stringify({
    action: "setSelfViewBackground",
    payloadJson: JSON.stringify(payload),
  });
}

export function sendSelfViewBackgroundToUnity(payload: UnitySelfViewBackgroundPayload): boolean {
  const ok = sendJsonToUnityReceiveController(buildSetSelfViewBackgroundCommandJson(payload));
  if (!ok) {
    console.warn("[unityBridge] sendSelfViewBackgroundToUnity failed: Unity not ready?", payload);
  }
  return ok;
}

export type UnitySpacePersonaSlot = {
  spacePersonaId: string;
  personaId: string;
  modelUrl: string;
  label?: string;
  personaKind?: string;
};

export type UnitySetSpacePersonasPayload = {
  spaceId: string;
  personas: UnitySpacePersonaSlot[];
  focusSpacePersonaId: string | null;
};

export type UnitySetFocusSpacePersonaPayload = {
  spacePersonaId: string | null;
};

export function buildSetSpacePersonasCommandJson(payload: UnitySetSpacePersonasPayload): string {
  return JSON.stringify({
    action: "setSpacePersonas",
    payloadJson: JSON.stringify(payload),
  });
}

export function buildSetFocusSpacePersonaCommandJson(
  payload: UnitySetFocusSpacePersonaPayload
): string {
  return JSON.stringify({
    action: "setFocusSpacePersona",
    payloadJson: JSON.stringify(payload),
  });
}

export function sendSpacePersonasToUnity(payload: UnitySetSpacePersonasPayload): boolean {
  const json = buildSetSpacePersonasCommandJson(payload);
  const ok = sendJsonToUnityReceiveController(json);
  if (!ok) {
    console.warn("[unityBridge] sendSpacePersonasToUnity failed: Unity not ready?", payload);
  } else {
    console.log(
      "[unityBridge] sendSpacePersonasToUnity:",
      payload.spaceId,
      payload.personas.length,
      "slots"
    );
  }
  return ok;
}

export function sendFocusSpacePersonaToUnity(spacePersonaId: string | null): boolean {
  const json = buildSetFocusSpacePersonaCommandJson({ spacePersonaId });
  const ok = sendJsonToUnityReceiveController(json);
  if (!ok) {
    console.warn("[unityBridge] sendFocusSpacePersonaToUnity failed: Unity not ready?", {
      spacePersonaId,
    });
  }
  return ok;
}

function sendHoverOutlineStyleToUnity(colorHtml: string, width: number): boolean {
  return sendJsonToUnityReceiveController(buildSetHoverOutlineStyleCommandJson(colorHtml, width));
}

export function sendPersonaHoverOutlineVisibleToUnity(visible: boolean): boolean {
  return sendHoverOutlineStyleToUnity(
    CHAT_SCENE_HOVER_OUTLINE_COLOR,
    visible ? CHAT_SCENE_HOVER_OUTLINE_WIDTH : 0
  );
}

export function sendJsonToUnityReceiveController(json: string): boolean {
  const contentWindow = unityIframe?.contentWindow as UnityFrameWindow | null;
  const unityInstance = contentWindow?.unityInstance;
  if (!unityInstance) {
    return false;
  }
  unityInstance.SendMessage("ReceiveController", "Execute", json);
  return true;
}

function sendExpressionToUnity(
  name: string,
  value = 1,
  spacePersonaId?: string | null
): boolean {
  return sendJsonToUnityReceiveController(
    buildSetExpressionCommandJson(name, value, spacePersonaId)
  );
}

function sendBlushToUnity(value = 1, spacePersonaId?: string | null): boolean {
  return sendJsonToUnityReceiveController(buildSetBlushCommandJson(value, spacePersonaId));
}

const COMPOSITE_EXPRESSIONS = {
  tsundere: { expression: "Angry" as const, blush: 0.8 },
  hanikami: { expression: "Happiness" as const, blush: 0.7 },
  terawari: { expression: "Joy" as const, blush: 0.6 },
  komaridere: { expression: "Sad" as const, blush: 0.8 },
} as const;

type CompositeExpressionKey = keyof typeof COMPOSITE_EXPRESSIONS;

function sendCompositeExpressionToUnity(
  key: CompositeExpressionKey,
  spacePersonaId?: string | null
): boolean {
  const { expression, blush } = COMPOSITE_EXPRESSIONS[key];
  const ok1 = sendExpressionToUnity(expression, 1, spacePersonaId);
  const ok2 = sendBlushToUnity(blush, spacePersonaId);
  return ok1 && ok2;
}

export function sendMotionExpressionToUnity(
  expression: string,
  spacePersonaId?: string | null
): boolean {
  const raw = expression.trim();
  if (!raw) return false;
  const norm = raw.toLowerCase();
  if (norm === "none") {
    return sendExpressionToUnity("None", 0, spacePersonaId) && sendBlushToUnity(0, spacePersonaId);
  }
  if (norm === "neutral" || norm === "kiss") {
    return sendExpressionToUnity("Neutral", 1, spacePersonaId) && sendBlushToUnity(0, spacePersonaId);
  }
  if (norm === "amae" || norm === "amae2") {
    return sendFacialParametersJsonToUnity(randomAmaeFacialPreset(), spacePersonaId);
  }
  const facialPreset = facialPresetForMotionExpression(norm);
  if (facialPreset) {
    return sendFacialParametersJsonToUnity(facialPreset, spacePersonaId);
  }
  if (norm === "hanikami") {
    return sendCompositeExpressionToUnity("hanikami", spacePersonaId);
  }
  if (norm === "tsundere") {
    return sendCompositeExpressionToUnity("tsundere", spacePersonaId);
  }
  if (norm === "komaridere") {
    return sendCompositeExpressionToUnity("komaridere", spacePersonaId);
  }
  return sendExpressionToUnity(raw, 1, spacePersonaId) && sendBlushToUnity(0, spacePersonaId);
}

function sendEyesDarkToUnity(value = 1, spacePersonaId?: string | null): boolean {
  return sendJsonToUnityReceiveController(buildSetEyesDarkCommandJson(value, spacePersonaId));
}

export function sendPlayMotionToUnity(
  name: string,
  loop = true,
  spacePersonaId?: string | null,
  semanticSegmentIndex?: number
): boolean {
  const resolved = resolvePlayMotionName(name, semanticSegmentIndex ?? 0);
  return sendJsonToUnityReceiveController(
    buildPlayMotionCommandJson(resolved, loop, spacePersonaId)
  );
}
