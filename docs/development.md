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

## mise の起動方式（shims 推奨）

本テンプレートでは `mise activate` ではなく **shims 方式** を推奨します。
shims を `PATH` に静的に追加するだけで、非対話シェル・IDE 統合・サブプロセス
（例: Docker build から呼ばれる `pnpm`、エディタからの lint 実行）でも常に
同じツールが解決されます。

### PowerShell（Windows）

`$PROFILE`（通常は `Documents\PowerShell\Microsoft.PowerShell_profile.ps1`）に追記:

```powershell
$miseShims = "$env:LOCALAPPDATA\mise\shims"
if ((Test-Path $miseShims) -and ($env:PATH -notlike "*$miseShims*")) {
    $env:PATH = "$miseShims;$env:PATH"
}
```

### Bash / Zsh

`~/.bashrc` または `~/.zshrc` に追記:

```bash
export PATH="$HOME/.local/share/mise/shims:$PATH"
```

`.mise.toml` に新しいツールを追加した後は `mise reshim` を実行してください。

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
mise run check            # CI と同じ検査（下の「CI との対応」を参照）
mise run db:up            # DB コンテナのみ起動
mise run db:migrate       # マイグレーション適用
mise run db:migrate:version   # 現在のバージョンと dirty フラグを表示
```

`make dev` や `make check` のように、同名の `make` ターゲットも使えます（`:` は `-` に読み替えます。例: `make db-migrate-version`）。
実体は `.mise.toml` のタスクです。

## CI との対応

CI（`.github/workflows/ci.yml`）は `.mise.toml` のタスクを呼び出すだけです。
ツールのバージョンも `.mise.toml` の `[tools]` から入れるため、ローカルの `mise run check` と同じ検査になります。

| 検査 | タスク | `mise run check` | CI |
| ---- | ------ | ---------------- | -- |
| golangci-lint（gofmt の検査を含む） | `lint:backend`・`lint:cli` | ○ | ○ |
| ESLint と `tsc --noEmit` | `lint:frontend` | ○ | ○ |
| markdownlint | `lint:markdown` | ○ | ○ |
| actionlint | `lint:actions` | ○ | ○ |
| govulncheck | `vuln:backend`・`vuln:cli` | ○ | ○ |
| Go のユニットテスト | `test:backend`・`test:cli` | ○ | ○（backend は下の統合テストに含めて実行） |
| backend の統合テストとカバレッジしきい値（80%） | `test:integration` | ○ | ○ |
| Vitest とカバレッジしきい値 | `test:frontend` | ○ | ○ |
| `scripts/` のテスト | `test:scripts` | ○ | ○ |
| フロントの本番ビルド | `build:frontend:dist` | ○ | ○ |
| `pnpm audit --prod` | `audit:frontend` | ― | ○（失敗しても警告のみ） |

`pnpm audit` だけは `mise run check` に含めていません。監査 API の障害や、自分の変更と関係のない新しい脆弱性の公開で失敗するためです。
手元で確認するときは `mise run audit:frontend` を実行してください。

gofmt 未適用のファイルがあると `lint:backend`・`lint:cli` が失敗します。`mise run fmt` で整形してください。

Go のバージョンは `.mise.toml` でパッチまで固定しています（`1.26` のような指定だと、mise は公開直後のリリースを選ばないため、修正版の取り込みが遅れます）。
`mise run vuln` が標準ライブラリの脆弱性を報告したら、`.mise.toml` の `go` を「Fixed in」に示されたバージョンへ上げてください。

## 統合テスト

`backend/` の `*_integration_test.go`（`//go:build integration`）は、実際の PostgreSQL に対して実行します。

```bash
mise run test:integration   # テスト用 DB の起動 → マイグレーション → 全テスト → カバレッジ検査
mise run db:test:down       # テスト用 DB を片付ける
```

- テスト用 DB は `docker-compose.yml` の `db-test`（ポート 5433、DB 名 `webapp_test`）です。テストがテーブルを空にするため、開発用の `db`（ポート 5432）とは別のコンテナにしています。
- `db-test` のデータはコンテナ内の tmpfs にあり、コンテナを作り直すと消えます。
- 別の PostgreSQL に対して実行したいときは、`TEST_DATABASE_URL` を設定します。その DB のテーブルは空になるので、開発や本番の DB を指定しないでください。
- `mise run test:backend` は統合テストを含みません。DB を使わずに短時間で確認したいときに使います。
- カバレッジのしきい値は 80% で、`backend/AGENTS.md` の目標と同じ値です。モック（`internal/mock`）とエントリポイント（`cmd`）は計測対象から外しています。

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

デプロイ手順、必要な GitHub Secrets / Variables、Google Cloud IAM、Artifact Registry、
Secret Manager、Cloud SQL 接続方式、frontend の `${PORT}` 対応については
[deployment.md](deployment.md) を参照してください。
