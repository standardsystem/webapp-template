# 開発ガイド

## 前提

- [mise](https://mise.jdx.dev/getting-started.html)
- Docker / Docker Compose（`mise run dev` 用）

## 初期セットアップ

```bash
mise trust && mise install
cp .env.example .env
mise run setup
```

`setup` は Go モジュール（backend / cli）の取得とフロントの `pnpm install` を行います。

## よく使うコマンド

```bash
mise run dev              # Docker Compose（API + フロント）
mise run dev:backend
mise run dev:frontend
mise run test             # backend + cli + frontend のテスト
mise run lint             # golangci + eslint + markdownlint 等
mise run fmt              # gofmt + Prettier + markdownlint --fix
mise run lint:markdown
mise run fmt:markdown
mise run check            # lint + test（CI 相当）
```

## Markdown

ルールは `.markdownlint-cli2.jsonc` です。`mise run lint:markdown` で検証、`fmt:markdown` で自動修正できるものを適用します。

## pre-commit（任意）

コミット前に Markdown などを整えたい場合:

```bash
pipx install pre-commit   # または mise で pipx 経由の pre-commit を有効化
pre-commit install
```

フック定義はリポジトリルートの `.pre-commit-config.yaml` です。

## Cloud Run

GitHub Actions（`deploy.yml`）から Cloud Run へデプロイする場合、リポジトリの Secrets に次を設定します。

| Secret | 説明 |
|--------|------|
| `GCP_PROJECT_ID` | Google Cloud プロジェクト ID |
| `GCP_WORKLOAD_IDENTITY_PROVIDER` | Workload Identity Federation のプロバイダー |
| `GCP_SERVICE_ACCOUNT` | デプロイ用サービスアカウント |

ローカルからの手動デプロイはルートの `Makefile` の `make deploy` を参照してください。
