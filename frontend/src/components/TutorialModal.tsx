"use client";

import { useEffect, useState } from "react";
import { createPortal } from "react-dom";
import {
	fetchTutorialSlides,
	setTutorialDismissed,
	tutorialImageSrc,
	type TutorialSlide,
} from "@/lib/tutorial";

type TutorialModalProps = {
	isOpen: boolean;
	onClose: () => void;
};

export function TutorialModal({ isOpen, onClose }: TutorialModalProps) {
	const [slides, setSlides] = useState<TutorialSlide[]>([]);
	const [index, setIndex] = useState(0);
	const [dontShowAgain, setDontShowAgain] = useState(false);
	const [loadError, setLoadError] = useState(false);

	useEffect(() => {
		if (!isOpen) return;
		let cancelled = false;
		setIndex(0);
		setDontShowAgain(false);
		setLoadError(false);
		void fetchTutorialSlides()
			.then((data) => {
				if (!cancelled) setSlides(data);
			})
			.catch(() => {
				if (!cancelled) setLoadError(true);
			});
		return () => {
			cancelled = true;
		};
	}, [isOpen]);

	if (!isOpen || typeof document === "undefined") return null;

	const total = slides.length;
	const slide = slides[index];
	const isFirst = index <= 0;
	const isLast = total > 0 && index >= total - 1;

	const close = () => {
		if (dontShowAgain) setTutorialDismissed(true);
		onClose();
	};

	const goPrev = () => {
		if (!isFirst) setIndex((i) => i - 1);
	};

	const goNext = () => {
		if (isLast) {
			close();
			return;
		}
		setIndex((i) => i + 1);
	};

	return createPortal(
		<div
			className="fixed inset-0 z-[110] flex items-center justify-center bg-black/55 p-3 backdrop-blur-[2px] sm:p-6"
			onClick={close}
			role="presentation"
		>
			<div
				className="flex max-h-[min(920px,94dvh)] w-full max-w-5xl flex-col overflow-hidden rounded-2xl bg-white shadow-2xl"
				onClick={(e) => e.stopPropagation()}
				role="dialog"
				aria-modal="true"
				aria-labelledby="tutorial-modal-title"
			>
				<header className="relative shrink-0 border-b border-violet-100 px-5 pb-4 pt-5 text-center sm:px-8 sm:pt-6">
					<button
						type="button"
						onClick={close}
						className="absolute right-3 top-3 flex h-9 w-9 items-center justify-center rounded-full text-2xl leading-none text-gray-400 hover:bg-gray-100 hover:text-gray-600 sm:right-4 sm:top-4"
						aria-label="閉じる"
					>
						×
					</button>
					<h2
						id="tutorial-modal-title"
						className="pr-8 text-xl font-bold text-violet-600 sm:text-2xl"
					>
						AISTalkへようこそ！
					</h2>
					<p className="mt-2 text-sm text-gray-600 sm:text-base">
						デモ版の基本的な使い方をご紹介します（約1分で完了します）
					</p>
				</header>

				<div className="min-h-0 flex-1 overflow-y-auto px-4 py-4 sm:px-6 sm:py-5">
					{loadError ? (
						<p className="py-12 text-center text-sm text-red-600">
							チュートリアルデータの読み込みに失敗しました。
						</p>
					) : !slide ? (
						<p className="py-12 text-center text-sm text-gray-500">読み込み中…</p>
					) : (
						<div className="grid gap-5 lg:grid-cols-[minmax(0,1.35fr)_minmax(0,1fr)] lg:items-stretch lg:gap-6">
							<div className="overflow-hidden rounded-xl border border-violet-100 bg-violet-50/40">
								{/* eslint-disable-next-line @next/next/no-img-element */}
								<img
									src={tutorialImageSrc(slide.id)}
									alt={`チュートリアル ${index + 1}`}
									className="h-full max-h-[min(420px,46dvh)] w-full object-contain object-center lg:max-h-[min(520px,58dvh)]"
								/>
							</div>

							<div className="flex flex-col">
								<span className="mb-3 inline-flex w-fit rounded-md bg-violet-600 px-2.5 py-1 text-xs font-semibold text-white">
									{index + 1}/{total}
								</span>
								<h3 className="text-lg font-bold leading-snug text-violet-700 sm:text-xl">
									{slide.title}
								</h3>
								<p className="mt-3 text-sm leading-relaxed text-gray-700 sm:text-[15px]">
									{slide.description}
								</p>
								{slide.note ? (
									<div className="mt-4 rounded-xl bg-violet-50 px-4 py-3">
										<p className="mb-1.5 flex items-center gap-1.5 text-sm font-semibold text-violet-700">
											<svg
												xmlns="http://www.w3.org/2000/svg"
												width="16"
												height="16"
												viewBox="0 0 24 24"
												fill="none"
												stroke="currentColor"
												strokeWidth="2"
												aria-hidden
											>
												<path d="M9 18h6" />
												<path d="M10 22h4" />
												<path d="M12 2a7 7 0 0 0-4 12.7V17h8v-2.3A7 7 0 0 0 12 2z" />
											</svg>
											ポイント
										</p>
										<p className="text-sm leading-relaxed text-violet-900/80">{slide.note}</p>
									</div>
								) : null}
							</div>
						</div>
					)}
				</div>

				<footer className="shrink-0 border-t border-violet-100 px-4 py-4 sm:px-6">
					<div className="flex items-center justify-between gap-3">
						<button
							type="button"
							onClick={goPrev}
							disabled={isFirst || !slide}
							className="rounded-lg border border-gray-200 bg-gray-50 px-3 py-2 text-sm font-medium text-gray-700 transition disabled:cursor-not-allowed disabled:opacity-40 hover:bg-gray-100 sm:px-4"
						>
							← 前へ
						</button>

						<div className="flex items-center gap-1.5" aria-label="進捗">
							{slides.map((s, i) => (
								<span
									key={s.id}
									className={`h-2.5 w-2.5 rounded-full transition-colors ${
										i === index ? "bg-violet-600" : "bg-violet-200"
									}`}
								/>
							))}
						</div>

						<button
							type="button"
							onClick={goNext}
							disabled={!slide}
							className="rounded-lg bg-violet-600 px-3 py-2 text-sm font-semibold text-white transition hover:bg-violet-700 disabled:opacity-40 sm:px-4"
						>
							{isLast ? "はじめる" : "次へ →"}
						</button>
					</div>

					<label className="mt-3 flex cursor-pointer items-center justify-center gap-2 text-sm text-gray-600">
						<input
							type="checkbox"
							checked={dontShowAgain}
							onChange={(e) => setDontShowAgain(e.target.checked)}
							className="h-4 w-4 rounded border-gray-300 accent-violet-600"
						/>
						次回から表示しない
					</label>
				</footer>
			</div>
		</div>,
		document.body,
	);
}
