export class CallOpenAIError extends Error {
	public override name: string = 'CallOpenAIError';
	public readonly status: number;

	constructor(response: Response) {
		const statusCode = response.status;
		const message = `ChatGPT API呼び出しエラーが発生しました。: ${statusCode}`;
		super(message);
		this.status = statusCode;
		Object.setPrototypeOf(this, CallOpenAIError.prototype);
	}
}
