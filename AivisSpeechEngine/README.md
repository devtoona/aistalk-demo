# AivisSpeech Engine

[AivisSpeech Engine](https://github.com/Aivis-Project/AivisSpeech-Engine) を Docker で起動します。  
公式イメージ [`ghcr.io/aivis-project/aivisspeech-engine`](https://github.com/Aivis-Project/AivisSpeech-Engine/pkgs/container/aivisspeech-engine) を使用します。

## 前提

- Docker（GPU 利用時は NVIDIA Container Toolkit）
- データ永続化用に `./data` をコンテナの `/home/user/.local/share/AivisSpeech-Engine-Dev` にマウントしています（[公式 README の Linux + Docker 手順](https://github.com/Aivis-Project/AivisSpeech-Engine) と同じパス）
- Linux で書き込みエラーになる場合は、`data` の所有者をコンテナユーザー（多くの環境で uid 1000）に合わせてください

## 起動

```bash
make pull          # 初回: CPU イメージ取得
make run           # CPU・フォアグラウンド（= make run-cpu）

make pull-gpu
make run-gpu       # GPU・フォアグラウンド

make run-cpu-bg    # CPU・バックグラウンド
make run-gpu-bg    # GPU・バックグラウンド

make stop          # バックグラウンド起動の停止
make logs          # バックグラウンド時のログ
```

## API

起動後 **`http://127.0.0.1:10101`** で HTTP API（VOICEVOX API 互換）。  
Swagger UI: http://127.0.0.1:10101/docs  

初回起動時はモデル・BERT のダウンロードで完了まで時間がかかることがあります。

## ローカルイメージをビルドする場合

公式イメージを `FROM` した薄い `Dockerfile` があります。

```bash
make build
# GPU ベースでビルドする例:
# docker build --build-arg VARIANT=nvidia-latest -t aistalk-aivis-speech-engine:local .
```

## aistalk バックエンドとの接続

ローカルエンジンを VOICEVOX 互換クライアントで叩く場合は、`VOICEVOX_API_URL=http://127.0.0.1:10101` のようにベース URL を 10101 に合わせます（話者は `/speakers` で取得した **style_id** を使用。Aivis では ID の扱いが VOICEVOX とは異なる点に注意）。
