# =============================================================================
# Makefile — mise run へのエイリアス
# 実体は .mise.toml の [tasks] に定義。make ユーザー向けの互換レイヤー。
# 推奨: mise run <task> を直接使用
# =============================================================================

.PHONY: setup dev dev-backend dev-frontend dev-down test test-backend test-frontend test-cli test-coverage lint lint-backend lint-frontend lint-cli fmt build clean check info db-up db-migrate db-status help

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

fmt:
	mise run fmt

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

db-status:
	mise run db:status

help:
	@echo "利用可能なタスク (mise run --list で詳細表示):"
	@mise tasks ls
