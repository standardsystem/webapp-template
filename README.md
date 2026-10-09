# webapp-template

Go + React の Web アプリに加え、**Go 製 CLI** を同じリポジトリで管理するモノレポテンプレートです。

**人向けの手順書・詳細ガイド:** [docs/README.md](docs/README.md)

**このテンプレートから新しい案件を始めるとき:** [docs/getting-started.md](docs/getting-started.md)

## 技術スタック

| 領域 | 技術 |
| ---- | --- |
| ツール管理 | mise（Go / Node と開発ツールのバージョン固定） |
| バックエンド | Go 1.26 / chi / distroless |
| CLI | Go 1.26（`cli/`・`go.work`） |
| フロントエンド | React 18 / Vite / TypeScript (strict) |
| パッケージ管理 | Go Modules / pnpm (Node) |
| タスクランナー | mise tasks（`mise run <task>`） |
| ドキュメント | `docs/`・markdownlint（`.markdownlint-cli2.jsonc`） |
| インフラ | Google Cloud Run / GitHub Actions |
| テスト（BE/CLI） | go test / table-driven |
| テスト（FE） | Vitest / Testing Library |
| ローカル | Docker Compose |

## 前提条件

- [mise](https://mise.jdx.dev/getting-started.html) がインストール済みであること
- Docker / Docker Compose

## クイックスタート

```bash
mise trust && mise install
cp .env.example .env
mise run setup
mise run dev
```

- フロントエンド: [http://localhost:5173](http://localhost:5173)（`/api` はバックエンドへ転送される）
- バックエンド: [http://localhost:8080](http://localhost:8080)
- ヘルス: [http://localhost:8080/health](http://localhost:8080/health)

## コマンド一覧

```bash
mise run setup           # 初期セットアップ（backend + cli + frontend）
mise run dev             # Docker Compose
mise run dev:backend
mise run dev:frontend
mise run test            # backend + cli + frontend
mise run test:cli
mise run lint            # Go（API/CLI）+ ESLint + markdownlint
mise run lint:markdown
mise run fmt             # gofmt + Prettier + markdownlint --fix
mise run fmt:markdown
mise run build
mise run test:integration  # backend の統合テスト（Docker でテスト用 DB を起動）
mise run check           # CI と同じ検査（lint + test + 統合テスト + govulncheck + フロントのビルド）
mise run db:migrate
mise run info

make dev
make check
```

## ディレクトリ構造

```text
.
├── docs/                 # 人向けマニュアル（本テンプレの「正」に近い文書群）
├── backend/              # Go API
├── cli/                  # Go CLI（例: version / health）
├── frontend/             # React + Vite
├── go.work               # backend / cli
├── .markdownlint-cli2.jsonc
├── .pre-commit-config.yaml
├── .cursor/
│   ├── rules/            # Markdown 規約（*.mdc）
│   └── skills/           # ドキュメント作業用スキル
├── docker-compose.yml
├── .mise.toml
├── Makefile
└── .github/workflows/
```

## アーキテクチャ（API）

```text
handler → usecase → domain ← repository
```

## CLI

```bash
cd cli && go run ./cmd/webapp-cli version
cd cli && go run ./cmd/webapp-cli health http://localhost:8080/health
```

## Cloud Run

GitHub の Secrets / Variables とデプロイ手順は [docs/deployment.md](docs/deployment.md)、DB マイグレーションは [docs/migration.md](docs/migration.md) を参照してください。

## AI との協働

- エージェント向けの指示は [AGENTS.md](AGENTS.md) と、各パッケージの `AGENTS.md`（[backend](backend/AGENTS.md) / [frontend](frontend/AGENTS.md) / [cli](cli/AGENTS.md)）
- `CLAUDE.md` は置かない。Claude Code は v2.1.281 以降が必要。構成の説明は [docs/ai-agents.md](docs/ai-agents.md)
