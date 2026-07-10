"use client";

/** ボタン 48px の外側にリングが乗るよう、中心 31・半径 26（内縁 > 24） */
const R = 26;
const CX = 31;
const CY = 31;
const VB = 62;
const CIRC = 2 * Math.PI * R;

type MicVolumeRingProps = {
	/** 0〜1 を円周の弧の長さにマッピング（内部で感度を上げて表示） */
	volume: number;
	className?: string;
};

/** RMS が小さくても弧が見えるよう表示用に持ち上げ（0〜1 にクランプ） */
function displayVolume(raw: number): number {
	const boosted = raw * 5.5 + 0.06;
	return Math.max(0, Math.min(1, boosted));
}

/**
 * マイク周りの円周バー（dB メーター的な伸び）
 */
export function MicVolumeRing({ volume, className = "" }: MicVolumeRingProps) {
	const v = displayVolume(volume);
	const dash = `${v * CIRC} ${CIRC}`;
	return (
		<svg
			className={`pointer-events-none absolute inset-0 h-full w-full -rotate-90 ${className}`.trim()}
			viewBox={`0 0 ${VB} ${VB}`}
			aria-hidden
		>
			{/* 枠（フルトラック） */}
			<circle cx={CX} cy={CY} r={R} fill="none" stroke="rgba(255,255,255,0.95)" strokeWidth="3" />
			{/* 音量バー */}
			<circle
				cx={CX}
				cy={CY}
				r={R}
				fill="none"
				stroke="rgba(0,0,0,0.88)"
				strokeWidth="3"
				strokeLinecap="round"
				strokeDasharray={dash}
				style={{ transition: "stroke-dasharray 90ms linear" }}
			/>
		</svg>
	);
}
