"use client";

type UnityTitleSplashProps = {
	/** Unity loader progress 0..1 */
	progress: number;
};

export function UnityTitleSplash({ progress }: UnityTitleSplashProps) {
	const pct = Math.max(0, Math.min(100, Math.round(progress * 100)));

	return (
		<div
			className="absolute inset-0 z-30 flex flex-col items-center justify-center bg-white"
			aria-busy="true"
			aria-live="polite"
		>
			{/* eslint-disable-next-line @next/next/no-img-element */}
			<img
				src="/branding/title-1024.png"
				alt="AISTalk"
				className="h-auto w-[min(72vw,420px)] max-w-[90vw] select-none object-contain"
				draggable={false}
			/>

			<div className="absolute bottom-[max(48px,12vh)] left-1/2 w-[min(420px,78vw)] -translate-x-1/2">
				<div
					className="h-2.5 w-full overflow-hidden rounded-full bg-violet-100"
					role="progressbar"
					aria-valuemin={0}
					aria-valuemax={100}
					aria-valuenow={pct}
					aria-label="読み込み中"
				>
					<div
						className="h-full rounded-full bg-gradient-to-r from-sky-400 via-indigo-500 to-violet-600 transition-[width] duration-150 ease-linear"
						style={{ width: `${pct}%` }}
					/>
				</div>
				<p className="mt-3 text-center text-sm font-medium tracking-wide text-gray-500">
					{pct}%
				</p>
			</div>
		</div>
	);
}
