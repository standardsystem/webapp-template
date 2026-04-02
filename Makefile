# =============================================================================
# Makefile — mise run へのエイリアス
# 実体は .mise.toml の [tasks] に定義。make ユーザー向けの互換レイヤー。
# 推奨: mise run <task> を直接使用
# =============================================================================

.PHONY: setup dev dev-down test test-coverage lint fmt build deploy clean check info help

## 初期セットアップ
setup:
	mise run setup

## ローカル開発
dev:
	mise run dev

dev-backend:
	mise run dev:backend

dev-frontend:
	mise run dev:frontend

dev-down:
	mise run dev:down

## テスト
test:
	mise run test

test-backend:
	mise run test:backend

test-frontend:
	mise run test:frontend

test-coverage:
	mise run test:coverage

## 静的解析
lint:
	mise run lint

lint-backend:
	mise run lint:backend

lint-frontend:
	mise run lint:frontend

## フォーマット
fmt:
	mise run fmt

## ビルド
build:
	mise run build

## Cloud Run デプロイ
deploy: build
	@echo "Cloud Run へデプロイ中..."
	gcloud run deploy webapp-template-api \
		--source=./backend \
		--region=asia-northeast1 \
		--allow-unauthenticated
	gcloud run deploy webapp-template-web \
		--source=./frontend \
		--region=asia-northeast1 \
		--allow-unauthenticated

## DB
db-up:
	mise run db:up

db-migrate:
	mise run db:migrate

db-rollback:
	mise run db:rollback

db-status:
	mise run db:status

## ユーティリティ
clean:
	mise run clean

check:
	mise run check

info:
	mise run info

## ヘルプ
help:
	@echo "利用可能なタスク (mise run --list で詳細表示):"
	@mise tasks ls
