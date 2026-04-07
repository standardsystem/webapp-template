# モノレポ構成

テンプレートは **API（Go）・フロント（React）・CLI（Go）** を1リポジトリで管理します。

## レイアウト

```text
.
├── backend/          # HTTP API（クリーンアーキテクチャ）
├── cli/              # コマンドラインツール（独立した Go モジュール）
├── frontend/         # React + Vite
├── docs/             # 人向けマニュアル
├── go.work           # Go ワークスペース（backend / cli）
├── .markdownlint-cli2.jsonc
└── .mise.toml
```

## Go ワークスペース

`go.work` で `backend` と `cli` をまとめています。ローカルではリポジトリルートで次のように扱えます。

```bash
go work sync   # 必要に応じて
```

各サブプロジェクトは **それぞれ `go test ./...`**、`golangci-lint run ./...`（作業ディレクトリを `backend` / `cli` に切り替え）で検証します。

## CLI のビルド

```bash
cd cli && go build -o webapp-cli ./cmd/webapp-cli
./webapp-cli version
```

## 運用上の境界

- **API と CLI** は別モジュール。共通ドメインを切り出す場合は `backend/internal` を直接 import せず、必要なら共有用の小さな内部モジュールや生成コードを別途検討します（テンプレート段階では未分離）。
- **ドキュメント** は `docs/` を正とし、ルートの `AGENTS.md` / `CLAUDE.md` はツール向けの要約＋リンクを残します。
