# AivisSpeech Engine
# https://github.com/Aivis-Project/AivisSpeech-Engine

IMAGE_REPO := ghcr.io/aivis-project/aivisspeech-engine
TAG_CPU := cpu-latest
TAG_GPU := nvidia-latest
PORT := 10101
CONTAINER_NAME := aivisspeech-engine
# モデル・BERT キャッシュ永続化（コンテナ内パスは公式 README 準拠）
DATA_DIR ?= $(abspath $(CURDIR)/data)

.PHONY: pull pull-gpu run run-cpu run-gpu run-cpu-bg run-gpu-bg stop logs build

# 公式イメージを取得（CPU）
pull:
	docker pull $(IMAGE_REPO):$(TAG_CPU)

pull-gpu:
	docker pull $(IMAGE_REPO):$(TAG_GPU)

# CPU で起動（フォアグラウンド）
run: run-cpu

run-cpu: $(DATA_DIR)
	docker run --rm -it \
		-p 127.0.0.1:$(PORT):$(PORT) \
		-v $(DATA_DIR):/home/user/.local/share/AivisSpeech-Engine-Dev \
		$(IMAGE_REPO):$(TAG_CPU)

# GPU で起動（フォアグラウンド）
run-gpu: $(DATA_DIR)
	docker run --rm -it --gpus all \
		-p 127.0.0.1:$(PORT):$(PORT) \
		-v $(DATA_DIR):/home/user/.local/share/AivisSpeech-Engine-Dev \
		$(IMAGE_REPO):$(TAG_GPU)

# CPU バックグラウンド
run-cpu-bg: $(DATA_DIR)
	docker run -d --name $(CONTAINER_NAME) \
		-p 127.0.0.1:$(PORT):$(PORT) \
		-v $(DATA_DIR):/home/user/.local/share/AivisSpeech-Engine-Dev \
		$(IMAGE_REPO):$(TAG_CPU)

# GPU バックグラウンド
run-gpu-bg: $(DATA_DIR)
	docker run -d --name $(CONTAINER_NAME) --gpus all \
		-p 127.0.0.1:$(PORT):$(PORT) \
		-v $(DATA_DIR):/home/user/.local/share/AivisSpeech-Engine-Dev \
		$(IMAGE_REPO):$(TAG_GPU)

stop:
	docker stop $(CONTAINER_NAME) 2>/dev/null || true
	docker rm $(CONTAINER_NAME) 2>/dev/null || true

logs:
	docker logs -f $(CONTAINER_NAME)

# Dockerfile からローカルタグを付与（VARIANT=nvidia-latest で GPU ベースも可）
build:
	docker build --build-arg VARIANT=$(TAG_CPU) -t aistalk-aivis-speech-engine:local .

$(DATA_DIR):
	mkdir -p $(DATA_DIR)
