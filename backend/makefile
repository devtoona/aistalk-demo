ifneq (,$(wildcard .env))
include .env
export
endif

ifndef APP_NAME
$(error APP_NAME is not set. Set it in .env or export APP_NAME)
endif

dev-build:
	docker build --no-cache -f Dockerfile.dev -t $(APP_NAME)-dev .

dev-run:
	docker run --rm -it -p 50037:50037 \
		--add-host=host.docker.internal:host-gateway \
		--env-file .env \
		$(APP_NAME)-dev

prod-build:
	docker build --no-cache -f Dockerfile.prod -t $(APP_NAME)-prod .

prod-run:
	docker run --rm -it -p 50037:50037 \
		--add-host=host.docker.internal:host-gateway \
		--env-file .env \
		$(APP_NAME)-prod

# PostgreSQL
db-up:
	docker compose up -d db

db-down:
	docker compose down

migrate-up:
	docker compose run --rm migrate -path /migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@db:5432/$(DB_NAME)?sslmode=disable" up

migrate-down:
	docker compose run --rm migrate -path /migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@db:5432/$(DB_NAME)?sslmode=disable" down 1

# テスト用初期データ投入（冪等、重複時はスキップ）. db が起動していること.
seed:
	docker compose exec -T db psql -U $(DB_USER) -d $(DB_NAME) < database/migrations/seed.sql

# Aivis 公開カタログ → voice_masters 同期（ホストで make → app コンテナ内で go run。.env の DB_HOST=db が効く）
# 例: make sync-aivis-voices ARGS=-dry-run  …取得・件数確認のみ（DB へ書かない）
# 実投入: make sync-aivis-voices（ARGS なし。db 起動済みであること）
# ホストで直接 go run したいときは DB_HOST=127.0.0.1 等にして手動実行。
sync-aivis-voices:
	docker compose run --rm app go run ./cmd/sync-aivis-voices $(ARGS)