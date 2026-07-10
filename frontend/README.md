# AISTalk Demo (Frontend)

VRM アバターと AI の会話デモ。チャット → モーション → Aivis 音声合成の最小構成。

## セットアップ

```bash
cp .env.example .env
npm install
npm run dev
```

- API: `NEXT_PUBLIC_API_BASE_URL`（例: `http://localhost:50037`）
- VRM: `public/models/avatar.vrm` に配置（`.gitignore` 対象）
- Unity WebGL: `public/unity/` に配置（`.gitignore` 対象）。本番デプロイは GCS から取得

## ビルド

```bash
npm run build   # out/ に静的 export
npm start       # out を 3000 で配信
```

## 機能

- 固定ペルソナ 1 体（`src/lib/demoConfig.ts` / 環境変数で上書き可）
- 会話履歴は `localStorage` に保存
- Firebase 匿名認証（設定時）+ Firestore 日次回数制限
- 課金・録画・管理画面なし
