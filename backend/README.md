# AISTalk Demo API (Go)

チャット・モーション推論・Aivis 音声合成の最小バックエンド。DB・認証なし。

## エンドポイント

| Method | Path | 説明 |
|--------|------|------|
| GET | `/healthz` | 死活監視 |
| POST | `/api/chat` | OpenAI チャット（`persona` を body で受け取る） |
| POST | `/api/avatar/motion` | モーション推論 |
| GET | `/api/event/stream/session/start?sessionId=...` | TTS 用 SSE |
| POST | `/api/event/tts/aivis/synthesize` | Aivis 音声合成 |

## 起動

```bash
cp .env.example .env
# OPENAI_API_KEY を設定

go run .
# http://localhost:50037
```

## Aivis Speech

- **local（デフォルト）**: [AivisSpeech Engine](https://github.com/Aivis-Project/AivisSpeech-Engine) を `127.0.0.1:10101` で起動
- **cloud**: `.env` に `AIVIS_MODE=cloud` と `AIVIS_API_KEY` を設定。`AIVIS_MODEL_UUID` でモデルを指定

## チャット API

```json
{
  "messages": [{ "role": "user", "content": "..." }],
  "space_persona_id": "demo-persona-1",
  "persona": {
    "name": "アシスタント",
    "personality": "明るく丁寧",
    "response_style": "共感的"
  }
}
```

## プロンプト

`prompts/files/single_chat_system_base.txt` と `avatar_motion_system_base.txt` を使用。`backend/` ディレクトリから起動すること。
