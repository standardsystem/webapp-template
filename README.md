# webapp-template

Go + React の Web アプリプロジェクトテンプレートです。Keep に蓄積したナレッジを反映した実践的な構成になっています。

## 技術スタック

| 領域 | 技術 |
|------|------|
| ツール管理 | mise（Go / Node / Python バージョン固定） |
| バックエンド | Go 1.26 / chi / distroless |
| フロントエンド | React 18 / Vite / TypeScript (strict) |
| パッケージ管理 | Go Modules / pnpm (Node) / uv (Python) |
| タスクランナー | mise tasks（`mise run <task>`） |
| インフラ | Google Cloud Run / GitHub Actions |
| テスト（BE） | go test / table-driven / mockgen |
| テスト（FE） | Vitest / Testing Library |
| ローカル | Docker Compose |

## 前提条件

- [mise](https://mise.jdx.dev/getting-started.html) がインストール済みであること
- Docker / Docker Compose

## クイックスタート

```bash
# 1. mise でツールをインストール
mise trust && mise install

# 2. 環境変数 & 依存パッケージ
cp .env.example .env
mise run setup          # go mod download + pnpm install

# 3. 起動（ホットリロード対応）
mise run dev

# 4. アクセス
# フロントエンド: http://localhost:5173
# バックエンド:   http://localhost:8080
# ヘルスチェック: http://localhost:8080/health
```

## コマンド一覧

```bash
mise run setup           # 初期セットアップ
mise run dev             # ローカル開発サーバー起動 (Docker Compose)
mise run dev:backend     # バックエンドのみ (air ホットリロード)
mise run dev:frontend    # フロントエンドのみ (Vite)
mise run test            # 全テスト実行
mise run test:coverage   # カバレッジレポート生成
mise run lint            # 静的解析（Go + TypeScript）
mise run fmt             # コードフォーマット（gofmt + Prettier）
mise run build           # Docker イメージビルド
mise run db:migrate      # DB マイグレーション
mise run check           # CI 相当（lint + test）
mise run clean           # クリーンアップ
mise run info            # 環境情報表示

# make エイリアスも利用可能（互換用）
make dev                 # → mise run dev
make deploy              # Cloud Run デプロイ
```

## ディレクトリ構造

```
.
├── CLAUDE.md              # Claude Code 向け設定
├── AGENTS.md              # AI エージェント共通設定
├── .env.example           # 環境変数テンプレート
├── .mise.toml             # mise: ツールバージョン & タスク定義
├── docker-compose.yml     # ローカル開発環境
├── Makefile               # mise run へのエイリアス（互換用）
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
