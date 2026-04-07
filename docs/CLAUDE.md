# CLAUDE.md — AI 協働開発ガイド

このファイルは Claude Code（および互換 AI エージェント）がプロジェクトを正しく理解して作業するための設定ファイルです。

ルートの `CLAUDE.md` は要約版です。矛盾がある場合は **このファイルを正** とします。

---

## プロジェクト概要

- **名称**: webapp-template（Web + CLI モノレポ）
- **構成**: Go バックエンド API + Go CLI + React (Vite + TypeScript)
- **ツール管理**: mise（Go / Node / Python のバージョン固定）
- **パッケージ管理**: Go Modules（`go.work`）/ pnpm (Node) / uv (Python)
- **タスクランナー**: mise tasks（`mise run <task>`）
- **ドキュメント**: 人向けマニュアルは `docs/`
- **デプロイ先**: Google Cloud Run（API / フロント）
- **CI/CD**: GitHub Actions

---

## ディレクトリ構造

```text
.
├── backend/          # Go API（クリーンアーキテクチャ）
│   ├── cmd/server/
│   └── internal/
│       ├── handler/
│       ├── usecase/
│       ├── domain/
│       └── repository/
├── cli/              # Go CLI（独立モジュール）
│   └── cmd/webapp-cli/
├── frontend/         # React + Vite + TypeScript
│   └── src/
│       ├── components/
│       ├── hooks/
│       └── lib/
├── docs/             # 人向けマニュアル
├── go.work           # backend / cli
├── .mise.toml
├── .github/workflows/
├── docker-compose.yml
└── Makefile
```

---

## 開発ルール

### アーキテクチャ（API）

- **依存方向**: handler → usecase → domain ← repository（依存は内側へ）
- **インターフェース**: usecase と repository の境界はインターフェースで定義する
- **ドメインロジック**: usecase/domain 以外にビジネスロジックを書かない

### Go コーディング規約

- `gofmt` / `golangci-lint` を必ず通す（`mise run lint`）
- エラーは必ず上位に返す（`fmt.Errorf("wrap: %w", err)` 形式）
- パッケージ名はシンプルに（例: `handler`, `usecase`, `domain`）
- テーブル駆動テストで書く（`t.Run` + サブテスト）
- API のモックは `internal/mock/` に配置

### TypeScript / React コーディング規約

- `strict: true` を維持する
- コンポーネントは関数コンポーネントのみ（クラスコンポーネント禁止）
- `any` 型の使用禁止（`unknown` を使う）
- カスタムフックでロジックを分離する
- テストは Vitest + Testing Library で書く

### テスト（TDD）

- **新機能**: テストを先に書いてから実装する
- **カバレッジ**: バックエンド 80% 以上、フロントエンド 70% 以上を目標
- **ハーネステスト**: AI が生成したコードは必ずテストで検証する
- テストファイルは実装ファイルと同じディレクトリに配置（`*_test.go` / `*.test.tsx`）

### コミット規約

```text
feat: 新機能
fix: バグ修正
docs: ドキュメント
test: テスト追加・修正
refactor: リファクタリング
chore: ビルド・設定変更
```

### Markdown

- ルールの参照元は `.markdownlint-cli2.jsonc`
- `MD013`（行長）はプロジェクト設定で無効

---

## セットアップ

```bash
# 1. mise インストール (https://mise.jdx.dev/getting-started.html)
# 2. プロジェクト初期化
mise trust && mise install
mise run setup              # Go (backend, cli) + pnpm install
```

## よく使うコマンド

```bash
mise run dev              # フロント・バック同時起動 (Docker Compose)
mise run dev:backend
mise run dev:frontend
mise run test             # backend + cli + frontend
mise run lint             # 静的解析（含: markdownlint）
mise run fmt              # gofmt + Prettier + markdown fix
mise run build
mise run check            # CI 相当 (lint + test)
mise run db:migrate
mise run info

make dev
make test
make deploy               # Cloud Run デプロイ
```

---

## AI へのお願い

1. **コードを生成したら必ずテストも生成する**
2. **既存のアーキテクチャの依存方向を崩さない**
3. **型安全を守る（Go の interface、TS の型定義）**
4. **エラーハンドリングを省略しない**
5. **大きな変更は小さなステップに分けて提案する**
6. **長い手順書は `docs/` に置く**
