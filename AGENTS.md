# AGENTS.md

Go API（`backend/`）+ Go CLI（`cli/`）+ React（`frontend/`）のモノレポテンプレートです。
このファイルは Claude Code / Codex / Cursor などが共通で読む、リポジトリ全体の指示です。
各パッケージ固有の規約は `backend/AGENTS.md`・`frontend/AGENTS.md`・`cli/AGENTS.md` にあります。

## 構成

```text
backend/    Go API（クリーンアーキテクチャ、独立モジュール）
cli/        Go CLI（独立モジュール）
frontend/   React + Vite + TypeScript
docs/       人向けマニュアル
go.work     backend / cli の Go ワークスペース
.mise.toml  ツールのバージョン固定とタスク定義
```

## コマンド

ツールとタスクは mise で管理します。リポジトリルートで実行してください。

```bash
mise run setup          # 初回セットアップ（Go モジュール取得 + pnpm install）
mise run dev            # Docker Compose で API + フロント + DB を起動
mise run test           # backend + cli + frontend のテスト
mise run lint           # golangci-lint + eslint + tsc + markdownlint
mise run fmt            # gofmt + Prettier + markdownlint --fix
mise run check          # lint + test（CI 相当）
mise run db:migrate     # DB マイグレーション適用
```

パッケージ単位のタスクは `test:backend`・`lint:frontend`・`test:cli` のように `:` で絞れます。
一覧は `.mise.toml` を参照してください。

## 必ず守ること

- 実装を追加・変更したら、同じ変更でテストも追加・更新する。
- 完了を報告する前に、触ったパッケージの lint とテストを実行して通す。
- エラーを握りつぶさない。Go では `_ = err` を書かず、上位へ返す。
- 既存のテストを削除・無効化しない。スキップする場合は理由をコードに書く。
- 秘密情報をコードやコミットに含めない。環境変数を追加したら `.env.example` にも追記する。
- `main` へ直接プッシュしない。ブランチを切って PR を出す。
- 大きな変更は小さなステップに分けて提案する。

## コミットと PR

- コミットは Conventional Commits 形式（`feat:` `fix:` `docs:` `test:` `refactor:` `chore:`）。
- ブランチ名は `feat/xxx`・`fix/xxx`・`docs/xxx` のように種別を先頭に付ける。
- PR は 400 行以内を目安にし、破壊的変更は説明に明記する。

## ドキュメント

- 人向けの長い手順書は `docs/` に置く。索引は `docs/README.md`。
- Markdown は `.markdownlint-cli2.jsonc` に従う（行長 `MD013` は無効）。編集後は `mise run lint:markdown` を通す。
- セットアップは `docs/development.md`、デプロイは `docs/deployment.md`、マイグレーションは `docs/migration.md`。
- 指示ファイルの置き方は `docs/ai-agents.md`。`CLAUDE.md` は作らない。このファイルは 200 行以内に保ち、パッケージ固有の内容は各パッケージの `AGENTS.md` に書く。
