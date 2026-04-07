# AGENTS.md — AI エージェント共通設定（要約）

> **完全版:** [docs/AGENTS.md](docs/AGENTS.md)（テスト例・API 規約・詳細マップ）

---

## Agent Identity

- Go API（`backend/`）・Go CLI（`cli/`）・React（`frontend/`）の実装とテスト
- CI/CD とハーネス（生成コードの自動検証）

詳細な責務は [docs/AGENTS.md](docs/AGENTS.md) を参照してください。

---

## Constraints（制約）

### やってはいけないこと

- `internal/domain/` 以外にビジネスロジックを書くこと
- テストなしで実装コードを追加すること
- `any` 型（TypeScript）や `interface{}` （Go）を安易に使うこと
- 秘密情報をコードに直書きすること
- `main` ブランチへの直接プッシュ
- 既存のテストを削除・無効化すること（スキップは理由を明記）

### 必ずやること

- 新しい関数・メソッドには必ずユニットテストを追加する
- エラーは握りつぶさない（Go: `_ = err` 禁止）
- 環境変数は `.env.example` に追記する
- 破壊的変更は PR の説明に明記する
- 人間向けの長文マニュアルは `docs/` に置く

---

## リポジトリの正規ドキュメント

- マニュアル索引: [docs/README.md](docs/README.md)
- Claude / 開発の詳細: [docs/CLAUDE.md](docs/CLAUDE.md)
- Markdown 規約: `.markdownlint-cli2.jsonc`（`MD013` は無効）

---

## Commit & PR（要約）

- ブランチ名: `feat/xxx`, `fix/xxx`, `test/xxx`
- レビュー前に `mise run check` を通す
