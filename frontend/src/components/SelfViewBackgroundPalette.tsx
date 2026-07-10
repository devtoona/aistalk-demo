"use client";

import type { SelfViewBackgroundPreset } from "@/lib/selfViewBackgroundPresets";
import {
	GRADIENT_BACKGROUND_PRESETS,
	SOLID_BACKGROUND_PRESETS,
} from "@/lib/selfViewBackgroundPresets";

type SelfViewBackgroundPaletteProps = {
	selectedId: string;
	onSelect: (preset: SelfViewBackgroundPreset) => void;
};

function SwatchButton({
	preset,
	selected,
	onSelect,
}: {
	preset: SelfViewBackgroundPreset;
	selected: boolean;
	onSelect: (preset: SelfViewBackgroundPreset) => void;
}) {
	return (
		<button
			type="button"
			title={preset.label}
			aria-label={`背景: ${preset.label}`}
			aria-pressed={selected}
			onClick={() => onSelect(preset)}
			className={`group flex flex-col items-center gap-1.5 rounded-lg p-1 transition-colors ${
				selected ? "bg-white/15 ring-1 ring-white/40" : "hover:bg-white/10"
			}`}
		>
			<span
				className="block h-10 w-10 rounded-full border border-white/20 shadow-inner"
				style={{
					background: preset.previewCss,
				}}
			/>
			<span className="max-w-[4.5rem] truncate text-[10px] text-white/70 group-hover:text-white/90">
				{preset.label}
			</span>
		</button>
	);
}

export function SelfViewBackgroundPalette({
	selectedId,
	onSelect,
}: SelfViewBackgroundPaletteProps) {
	return (
		<div className="space-y-4">
			<section>
				<p className="mb-2 text-xs font-semibold tracking-wide text-white/55">単色</p>
				<div className="grid grid-cols-4 gap-1">
					{SOLID_BACKGROUND_PRESETS.map((preset) => (
						<SwatchButton
							key={preset.id}
							preset={preset}
							selected={selectedId === preset.id}
							onSelect={onSelect}
						/>
					))}
				</div>
			</section>
			<section>
				<p className="mb-2 text-xs font-semibold tracking-wide text-white/55">グラデーション</p>
				<div className="grid grid-cols-4 gap-1">
					{GRADIENT_BACKGROUND_PRESETS.map((preset) => (
						<SwatchButton
							key={preset.id}
							preset={preset}
							selected={selectedId === preset.id}
							onSelect={onSelect}
						/>
					))}
				</div>
			</section>
		</div>
	);
}
