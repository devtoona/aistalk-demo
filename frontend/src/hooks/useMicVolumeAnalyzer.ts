"use client";

import { useCallback, useEffect, useRef, useState } from 'react';

export function useMicVolumeAnalyzer() {
	const [volume, setVolume] = useState(0);
	const [isActive, setIsActive] = useState(false);
	const isActiveRef = useRef(false);
	const audioCtxRef = useRef<AudioContext | null>(null);
	const micStreamRef = useRef<MediaStream | null>(null);
	/** false のときストリームは VAD 等の共有（stop しない） */
	const micStreamOwnedRef = useRef(true);
	const animationIdRef = useRef<number | null>(null);
	const analyserRef = useRef<AnalyserNode | null>(null);
	const isStartingRef = useRef(false);

	function calculateRMSVolume(data: Uint8Array): number {
		let sum = 0;
		for (let i = 0; i < data.length; i++) {
			const deviation = data[i] - 128;
			sum += deviation * deviation;
		}
		const rms = Math.sqrt(sum / data.length);
		return Math.min(1.0, rms / 128);
	}

	const startMicVolume = useCallback(async (existingStream?: MediaStream | null) => {
		if (isStartingRef.current || isActiveRef.current) return;
		isStartingRef.current = true;
		try {
			audioCtxRef.current = new (window.AudioContext ||
				(window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext)();
			if (audioCtxRef.current.state === 'suspended') {
				await audioCtxRef.current.resume();
			}
			const live =
				existingStream?.getAudioTracks().some((t) => t.readyState === 'live') ?? false;
			if (live && existingStream) {
				micStreamRef.current = existingStream;
				micStreamOwnedRef.current = false;
			} else {
				micStreamRef.current = await navigator.mediaDevices.getUserMedia({ audio: true });
				micStreamOwnedRef.current = true;
			}
			const source = audioCtxRef.current.createMediaStreamSource(micStreamRef.current);

			analyserRef.current = audioCtxRef.current.createAnalyser();
			analyserRef.current.fftSize = 512;
			const dataArray = new Uint8Array(analyserRef.current.frequencyBinCount);

			source.connect(analyserRef.current);
			// Safari 等: Analyser だけではタイムドメインが進まないことがある → 無音で destination まで繋ぐ
			const silent = audioCtxRef.current.createGain();
			silent.gain.value = 0;
			analyserRef.current.connect(silent);
			silent.connect(audioCtxRef.current.destination);

			const update = () => {
				if (analyserRef.current) {
					analyserRef.current.getByteTimeDomainData(dataArray);
					const vol = calculateRMSVolume(dataArray);
					setVolume(vol);
					animationIdRef.current = requestAnimationFrame(update);
				}
			};

			setIsActive(true);
			isActiveRef.current = true;
			update();
		} catch (error) {
			console.error('マイク音量解析開始エラー:', error);
			throw error;
		} finally {
			isStartingRef.current = false;
		}
	}, []);

	const stopMicVolume = useCallback(() => {
		if (animationIdRef.current) {
			cancelAnimationFrame(animationIdRef.current);
			animationIdRef.current = null;
		}
		if (micStreamRef.current && micStreamOwnedRef.current) {
			micStreamRef.current.getTracks().forEach((track) => track.stop());
		}
		micStreamRef.current = null;
		micStreamOwnedRef.current = true;
		if (audioCtxRef.current) {
			audioCtxRef.current.close();
			audioCtxRef.current = null;
		}
		setIsActive(false);
		isActiveRef.current = false;
		setVolume(0);
	}, []);

	useEffect(() => () => { stopMicVolume(); }, [stopMicVolume]);

	return {
		volume,
		isActive,
		startMicVolume,
		stopMicVolume,
	};
}
