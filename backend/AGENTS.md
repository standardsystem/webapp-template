# backend — Go API

`backend/` で作業するときの規約です。リポジトリ共通の指示はルートの `AGENTS.md` にあります。

## レイヤー

```text
cmd/server/                HTTP サーバーのエントリ
cmd/migrate/               マイグレーション実行用コマンド
internal/domain/           エンティティとリポジトリインターフェース
internal/usecase/          ビジネスロジック（domain のみに依存）
internal/handler/          HTTP 入出力とミドルウェア（usecase のみに依存）
internal/repository/       DB 実装（domain のインターフェースを実装）
internal/infrastructure/   DB 接続・OAuth プロバイダ・セッション
internal/mock/             テスト用モック（手書き）
migrations/                golang-migrate の SQL（embed で同梱）
```

- 依存方向は `handler → usecase → domain ← repository`。内側から外側を import しない。
- ビジネスロジックは `usecase` と `domain` にだけ書く。`handler` と `repository` には書かない。
- `usecase` と `repository` の境界は `domain` のインターフェースで定義する。

## コーディング

- エラーは `fmt.Errorf("...: %w", err)` でラップして上位へ返す。
- `interface{}` / `any` を安易に使わず、具体的な型を定義する。
- パッケージ名は短い単語にする（`handler`・`usecase`・`domain`）。
- API のパスは `/api/v1/{resource}`。レスポンスは `internal/handler/response.go` のヘルパーで JSON を返す。

## テスト

- テストは実装と同じディレクトリに `*_test.go` で置き、テーブル駆動（`t.Run` のサブテスト）で書く。
- 外部依存は `internal/mock/` のモックに差し替える。インターフェースを追加したらモックも追加する。
- `repository` など DB を使うテストは、ファイルの先頭に `//go:build integration` を付け、`*_integration_test.go` に置く。
- カバレッジは 80% 以上を保つ。`mise run test:integration` と CI が同じしきい値で検査する（`internal/mock` と `cmd` は計測対象外）。

```bash
mise run test:backend       # ユニットテストのみ（DB 不要）
mise run test:integration   # ユニット + 統合テスト + カバレッジ 80% の検査（Docker でテスト用 DB を起動）
mise run lint:backend       # golangci-lint run ./...
```

## マイグレーション

- `migrations/` に `NNNNNN_name.up.sql` と `.down.sql` を対で追加する。
- マージ済みのマイグレーションは書き換えず、新しい番号で追加する。
- 適用は `mise run db:migrate`、1 つ戻すのは `mise run db:migrate:down`。詳細は `docs/migration.md`。
