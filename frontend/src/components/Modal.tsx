"use client";

import React from "react";
import { createPortal } from "react-dom";

interface ModalProps {
	isOpen: boolean;
	onClose: () => void;
	title?: string;
	titleClassName?: string;
	/** ヘッダー左（戻るなど）。× は右端のまま */
	headerStart?: React.ReactNode;
	/** タイトル行の直下（区切りの下）。キャラクター一覧の戻るなど */
	belowHeader?: React.ReactNode;
	children: React.ReactNode;
	/** パネル幅（Tailwind）。既定: max-w-3xl */
	panelMaxWidthClass?: string;
	/** パネル縦上限（Tailwind）。既定: max-h-[90vh] */
	panelMaxHeightClass?: string;
	/**
	 * light: 従来のグレー紙面（フォーム向け）。
	 * dark: チャット入力バー寄りのニュートラルパネル。背面は暗転せず、軽いぼかしのみ。
	 */
	appearance?: "light" | "dark";
	/** true のときビューポートいっぱい（試験用・一覧向け） */
	fullscreen?: boolean;
	/**
	 * appearance=dark のとき、パネルを #09090b → neutral-600 のグラデにする（試験用）。
	 */
	panelGradientDark?: boolean;
}

export const Modal = ({
	isOpen,
	onClose,
	title,
	titleClassName,
	headerStart,
	belowHeader,
	children,
	panelMaxWidthClass = "max-w-3xl",
	panelMaxHeightClass = "max-h-[90vh]",
	appearance = "light",
	fullscreen = false,
	panelGradientDark = false,
}: ModalProps) => {
	if (!isOpen) return null;

	const isDark = appearance === "dark";
	/** dark は暗転なし＋背面を軽くぼかす。light は暗転＋同程度のぼかし */
	const overlayClass = isDark
		? "bg-transparent backdrop-blur-[3px]"
		: "bg-black/50 backdrop-blur-[3px]";
	/** 入力バーと同系の単色、または zinc 寄りからのグラデ（試験用） */
	const panelClass = (() => {
		if (!isDark) {
			return fullscreen
				? "border-0 bg-gray-100 p-6 shadow-none"
				: "bg-gray-100 p-6 shadow-xl";
		}
		const bg = panelGradientDark
			? "bg-gradient-to-br from-[#09090b] to-neutral-600"
			: "bg-neutral-600";
		if (fullscreen) {
			return `border-0 ${bg} p-6 shadow-none`;
		}
		return `border border-neutral-900/45 ${bg} p-6 shadow-[0_3px_14px_rgba(0,0,0,0.22)]`;
	})();
	const titleFallback = isDark ? "text-xl text-white" : "text-xl text-gray-900";
	const closeClass = isDark
		? "text-2xl text-white hover:text-white/85"
		: "text-2xl text-gray-500 hover:text-gray-700";

	const panelSizeClass = fullscreen
		? "min-h-[100dvh] h-[100dvh] max-h-[100dvh] w-full max-w-full overflow-y-auto rounded-none"
		: `${panelMaxHeightClass} w-full ${panelMaxWidthClass} overflow-y-auto rounded-xl`;

	const outerLayoutClass = fullscreen
		? "items-stretch justify-stretch p-0"
		: "items-center justify-center p-4";

	const showTitleBar = Boolean(title || headerStart);
	const titleBarClass =
		!showTitleBar
			? "mb-0"
			: belowHeader
				? `flex items-center gap-2 sm:gap-3 ${isDark ? "border-b border-white/20 pb-3" : "border-b border-gray-200 pb-3"}`
				: `flex items-center gap-2 sm:gap-3 ${isDark ? "mb-4 border-b border-white/20 pb-3" : "mb-4"}`;

	const belowHeaderWrapClass = belowHeader ? "mb-4 mt-2" : "";

	const modalContent = (
		<div
			className={`fixed inset-0 z-[100] flex ${outerLayoutClass} ${overlayClass}`}
			onClick={onClose}
			role="presentation"
		>
			<div
				className={`${panelSizeClass} ${panelClass}`}
				onClick={(e) => e.stopPropagation()}
				role="dialog"
				aria-modal="true"
				aria-labelledby={title ? "modal-title" : undefined}
			>
				<div className={titleBarClass}>
					{headerStart ? <div className="shrink-0">{headerStart}</div> : null}
					{title ? (
						<h2
							id="modal-title"
							className={`min-w-0 flex-1 font-bold ${titleClassName ?? titleFallback}`}
						>
							{title}
						</h2>
					) : (
						<span className="min-w-0 flex-1" />
					)}
					<button
						type="button"
						onClick={onClose}
						className={`shrink-0 focus:outline-none focus:ring-0 ${closeClass}`}
						aria-label="閉じる"
					>
						×
					</button>
				</div>
				{belowHeader ? <div className={belowHeaderWrapClass}>{belowHeader}</div> : null}
				{children}
			</div>
		</div>
	);

	if (typeof document === "undefined") return null;
	return createPortal(modalContent, document.body);
};
