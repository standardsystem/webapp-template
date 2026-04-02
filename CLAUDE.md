# CLAUDE.md — AI 協働開発ガイド

このファイルは Claude Code（および互換 AI エージェント）がプロジェクトを正しく理解して作業するための設定ファイルです。

---

## プロジェクト概要

- **名称**: webapp-template
- **構成**: Go バックエンド API + React (Vite + TypeScript) フロントエンド
- **ツール管理**: mise（Go / Node / Python のバージョン固定）
- **パッケージ管理**: Go Modules / pnpm (Node) / uv (Python)
- **タスクランナー**: mise tasks（`mise run <task>`）
- **デプロイ先**: Google Cloud Run
- **CI/CD**: GitHub Actions

---

## ディレクトリ構造

```
.
├── backend/          # Go API サーバー（クリーンアーキテクチャ）
│   ├── cmd/server/   # エントリポイント
│   └── internal/
│       ├── handler/  # HTTP ハンドラ層（外側）
│       ├── usecase/  # ユースケース層（ビジネスロジック）
│       ├── domain/   # ドメイン層（エンティティ・インターフェース）
│       └── repository/ # データアクセス層
├── frontend/         # React + Vite + TypeScript
│   └── src/
│       ├── components/ # UI コンポーネント
│       ├── hooks/      # カスタムフック
│       └── lib/        # ユーティリティ
├── .mise.toml         # mise: ツールバージョン & タスク定義
├── .github/workflows/ # CI/CD パイプライン
├── docker-compose.yml # ローカル開発環境
└── Makefile           # mise run へのエイリアス（互換用）
```

---

## 開発ルール

### アーキテクチャ
- **依存方向**: handler → usecase → domain ← repository（依存は内側へ）
- **インターフェース**: usecase と repository の境界はインターフェースで定義する
- **ドメインロジック**: usecase/domain 以外にビジネスロジックを書かない

### Go コーディング規約
- `gofmt` / `golangci-lint` を必ず通す（`mise run lint`）
- エラーは必ず上位に返す（`fmt.Errorf("wrap: %w", err)` 形式）
- パッケージ名はシンプルに（例: `handler`, `usecase`, `domain`）
- テーブル駆動テストで書く（`t.Run` + サブテスト）
- モックは `internal/mock/` に配置

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
```
feat: 新機能
fix: バグ修正
docs: ドキュメント
test: テスト追加・修正
refactor: リファクタリング
chore: ビルド・設定変更
```

---

## セットアップ

```bash
# 1. mise インストール (https://mise.jdx.dev/getting-started.html)
# 2. プロジェクト初期化
mise trust && mise install
mise run setup              # Go mod download + pnpm install
```

## よく使うコマンド

```bash
# タスクランナー（推奨）
mise run dev              # フロント・バック同時起動 (Docker Compose)
mise run dev:backend      # バックエンドのみ (air ホットリロード)
mise run dev:frontend     # フロントエンドのみ (Vite)
mise run test             # 全テスト実行
mise run lint             # 静的解析
mise run fmt              # コードフォーマット
mise run build            # 本番ビルド
mise run check            # CI 相当 (lint + test)
mise run db:migrate       # DB マイグレーション
mise run info             # 環境情報表示

# make エイリアス（互換用）
make dev                  # → mise run dev
make test                 # → mise run test
make deploy               # Cloud Run デプロイ
```

---

## AI へのお願い

1. **コードを生成したら必ずテストも生成する**
2. **既存のアーキテクチャの依存方向を崩さない**
3. **型安全を守る（Go の interface、TS の型定義）**
4. **エラーハンドリングを省略しない**
5. **大きな変更は小さなステップに分けて提案する**
