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
