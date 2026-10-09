# =============================================================================
# Makefile — mise run へのエイリアス
# 実体は .mise.toml の [tasks] に定義。make ユーザー向けの互換レイヤー。
# 推奨: mise run <task> を直接使用
# =============================================================================

.PHONY: setup dev dev-backend dev-frontend dev-down test test-backend test-frontend test-cli test-integration test-coverage lint lint-backend lint-frontend lint-cli lint-markdown fmt fmt-markdown build clean check info db-up db-migrate db-migrate-down db-migrate-version help

setup:
	mise run setup

dev:
	mise run dev

dev-backend:
	mise run dev:backend

dev-frontend:
	mise run dev:frontend

dev-down:
	mise run dev:down

test:
	mise run test

test-backend:
	mise run test:backend

test-frontend:
	mise run test:frontend

test-cli:
	mise run test:cli

test-integration:
	mise run test:integration

test-coverage:
	mise run test:coverage

lint:
	mise run lint

lint-backend:
	mise run lint:backend

lint-frontend:
	mise run lint:frontend

lint-cli:
	mise run lint:cli

lint-markdown:
	mise run lint:markdown

fmt:
	mise run fmt

fmt-markdown:
	mise run fmt:markdown

build:
	mise run build

clean:
	mise run clean

check:
	mise run check

info:
	mise run info

db-up:
	mise run db:up

db-migrate:
	mise run db:migrate

db-migrate-down:
	mise run db:migrate:down

db-migrate-version:
	mise run db:migrate:version

help:
	@echo "利用可能なタスク (mise run --list で詳細表示):"
	@mise tasks ls
