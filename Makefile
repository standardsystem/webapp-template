.PHONY: dev test lint build deploy clean help

## ローカル開発
dev:
	docker compose up --build --watch

dev-down:
	docker compose down

## テスト
test:
	@echo "=== Backend Tests ==="
	cd backend && go test -v -race -cover ./...
	@echo "=== Frontend Tests ==="
	cd frontend && npm run test

test-coverage:
	@echo "=== Backend Coverage ==="
	cd backend && go test -race -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html
	@echo "=== Frontend Coverage ==="
	cd frontend && npm run test:coverage

## 静的解析
lint:
	@echo "=== Backend Lint ==="
	cd backend && golangci-lint run ./...
	@echo "=== Frontend Lint ==="
	cd frontend && npm run lint && npm run type-check

## ビルド
build:
	@echo "=== Backend Build ==="
	cd backend && docker build -t webapp-backend .
	@echo "=== Frontend Build ==="
	cd frontend && docker build -t webapp-frontend .

## Cloud Run デプロイ（ローカルから手動実行用）
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

## 初期セットアップ
setup:
	cd frontend && npm install
	cd backend && go mod download

## クリーンアップ
clean:
	docker compose down -v
	cd backend && rm -f coverage.out coverage.html
	cd frontend && rm -rf dist coverage

## ヘルプ
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
