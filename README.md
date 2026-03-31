# webapp-template

Go + React の Web アプリプロジェクトテンプレートです。Keep に蓄積したナレッジを反映した実践的な構成になっています。

## 技術スタック

| 領域 | 技術 |
|------|------|
| バックエンド | Go 1.22 / chi / distroless |
| フロントエンド | React 18 / Vite / TypeScript (strict) |
| インフラ | Google Cloud Run / GitHub Actions |
| テスト（BE） | go test / table-driven / mockgen |
| テスト（FE） | Vitest / Testing Library |
| ローカル | Docker Compose |

## クイックスタート

```bash
# 1. セットアップ
cp .env.example .env
make setup

# 2. 起動（ホットリロード対応）
make dev

# 3. アクセス
# フロントエンド: http://localhost:5173
# バックエンド:   http://localhost:8080
# ヘルスチェック: http://localhost:8080/health
```

## コマンド一覧

```bash
make dev             # ローカル開発サーバー起動
make test            # 全テスト実行
make test-coverage   # カバレッジレポート生成
make lint            # 静的解析（Go + TypeScript）
make build           # Docker イメージビルド
make deploy          # Cloud Run デプロイ（手動）
make clean           # クリーンアップ
```

## ディレクトリ構造

```
.
├── CLAUDE.md              # Claude Code 向け設定
├── AGENTS.md              # AI エージェント共通設定
├── .env.example           # 環境変数テンプレート
├── docker-compose.yml     # ローカル開発環境
├── Makefile               # 開発コマンド集
│
├── backend/               # Go API サーバー
│   ├── cmd/server/        # エントリポイント
│   ├── internal/
│   │   ├── handler/       # HTTP 層（テスト済み）
│   │   ├── usecase/       # ビジネスロジック層（テスト済み）
│   │   ├── domain/        # エンティティ・インターフェース
│   │   └── repository/    # DB アクセス層
│   └── Dockerfile         # マルチステージ / distroless
│
├── frontend/              # React + Vite + TypeScript
│   ├── src/
│   │   ├── components/    # UI コンポーネント（テスト済み）
│   │   ├── hooks/         # カスタムフック
│   │   └── lib/           # API クライアント・ユーティリティ
│   └── Dockerfile         # マルチステージ / nginx
│
└── .github/workflows/
    ├── ci.yml             # PR ごとにテスト・lint
    └── deploy.yml         # main push → Cloud Run デプロイ
```

## アーキテクチャ

クリーンアーキテクチャ（依存は内側へ）を採用：

```
handler → usecase → domain ← repository
```

- **domain**: エンティティとインターフェース定義。外部依存なし
- **usecase**: ビジネスロジック。domain のみ依存
- **handler**: HTTP 入出力。usecase のみ依存
- **repository**: DB・外部 API の実装。domain インターフェースを実装

## Cloud Run デプロイ設定

GitHub Secrets に以下を設定してください：

| Secret | 説明 |
|--------|------|
| `GCP_PROJECT_ID` | Google Cloud プロジェクト ID |
| `GCP_WORKLOAD_IDENTITY_PROVIDER` | Workload Identity Federation のプロバイダー |
| `GCP_SERVICE_ACCOUNT` | デプロイ用サービスアカウント |

## Claude Code での開発

このテンプレートは Claude Code との協働を前提に設計されています：

- **CLAUDE.md** — プロジェクト構造・ルール・よく使うコマンドを記述
- **AGENTS.md** — AI エージェント向けの制約・テスト戦略・API 規約を記述
- テストを先に書いてから実装する TDD フローを推奨
- ハーネスエンジニアリング（AI 生成コードの自動検証）を CI で実施
