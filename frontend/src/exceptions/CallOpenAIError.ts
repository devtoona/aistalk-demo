export class CallOpenAIError extends Error {
	public override name: string = "CallOpenAIError";
	public readonly status: number;
	public readonly code: string | null;

	constructor(response: Response, code: string | null = null) {
		const statusCode = response.status;
		const message =
			code === "quota_exceeded"
				? "デモ版の利用上限に達しました。"
				: `ChatGPT API呼び出しエラーが発生しました。: ${statusCode}`;
		super(message);
		this.status = statusCode;
		this.code = code;
		Object.setPrototypeOf(this, CallOpenAIError.prototype);
	}

	get isQuotaExceeded(): boolean {
		return this.status === 429 || this.code === "quota_exceeded";
	}
}

export async function errorFromFailedResponse(response: Response): Promise<CallOpenAIError> {
	let code: string | null = null;
	try {
		const body = (await response.clone().json()) as { error?: string };
		if (typeof body?.error === "string") code = body.error;
	} catch {
		// ignore non-JSON
	}
	return new CallOpenAIError(response, code);
}
