# cli — Go CLI

`cli/` で作業するときの規約です。リポジトリ共通の指示はルートの `AGENTS.md` にあります。

- エントリは `cmd/webapp-cli/`。サブコマンドは 1 ファイル 1 コマンドで追加する（例: `health.go`・`version.go`）。
- `cli` は `backend` と別の Go モジュール。`backend/internal` を直接 import しない。
- エラーは `fmt.Errorf("...: %w", err)` でラップして返し、終了コードで失敗を伝える。
- テストは `*_test.go` にテーブル駆動で書く。

```bash
mise run test:cli   # go test -v -race -cover ./...
mise run lint:cli   # golangci-lint run ./...
```
