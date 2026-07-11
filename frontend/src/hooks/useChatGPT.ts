"use client";

import { CallOpenAIError } from '@/exceptions/CallOpenAIError';
import { useTTSLoading } from '@/contexts/TTSLoadingContext';
import { authFetch } from '@/lib/apiClient';
import type { FetchOpenAIResponse } from '@/types/FetchOpenAIResponse';

const API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL || '';

export type ChatParticipantPayload = {
	space_persona_id: string;
	label?: string;
	persona_kind?: string;
};

type PersonaPayload = {
	name?: string;
	personality?: string;
	response_style?: string;
	tempo?: string;
};

type ChatContext = {
	space_persona_id?: string;
	participants?: ChatParticipantPayload[];
	persona?: PersonaPayload;
};

export function useChatGPT() {
	const { setTTSLoading } = useTTSLoading();

	const sendMessage = async (messages: unknown[], context?: ChatContext) => {
		setTTSLoading(true);
		try {
			const body: Record<string, unknown> = { messages };
			if (context?.persona) body.persona = context.persona;
			if (context?.space_persona_id !== undefined) body.space_persona_id = context.space_persona_id;
			if (context?.participants !== undefined) body.participants = context.participants;

			const response = await authFetch(`${API_BASE}/api/chat`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(body),
			});

			if (!response.ok) {
				throw new CallOpenAIError(response);
			}

			const data = await response.json();
			return data as FetchOpenAIResponse;
		} catch (error) {
			console.error('ChatGPT APIエラー:', error);
			throw error;
		} finally {
			console.log('ChatGPT API呼び出し完了');
		}
	};

	return { sendMessage };
}
