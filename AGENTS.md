# AGENTS.md — AI エージェント共通設定

> このファイルは Claude Code / Codex / Cursor など複数の AI エージェントが共通で参照する設定ファイルです。
> 60,000+ リポジトリで採用されている事実上の共通フォーマットに準拠しています。

---

## Agent Identity

このリポジトリで動作する AI エージェントは以下の役割を担います：

- Go バックエンド API の実装・テスト・リファクタリング
- React フロントエンドのコンポーネント実装・テスト
- CI/CD パイプラインの維持・改善
- コードレビューの補助（ハーネスエンジニアリング）

---

## Constraints（制約）

### やってはいけないこと

- `internal/domain/` 以外にビジネスロジックを書くこと
- テストなしで実装コードを追加すること
- `any` 型（TypeScript）や `interface{}` （Go）を安易に使うこと
- 秘密情報（APIキー・パスワード）をコードに直書きすること
- `main` ブランチへの直接プッシュ
- 既存のテストを削除・無効化すること（スキップは理由を明記）

### 必ずやること

- 新しい関数・メソッドには必ずユニットテストを追加する
- エラーは握りつぶさない（Go: `_ = err` 禁止）
- 環境変数は `.env.example` に追記する
- 破壊的変更は PR の説明に明記する

---

## Project Structure Map

```
backend/internal/
  domain/       ← エンティティ・リポジトリインターフェース定義
  usecase/      ← ビジネスロジック（domain のみ依存可）
  handler/      ← HTTP 入出力（usecase のみ依存可）
  repository/   ← DB・外部API実装（domain インターフェースを実装）
  mock/         ← テスト用モック（自動生成: mockgen）

frontend/src/
  components/   ← UI コンポーネント（ロジックを持たない）
  hooks/        ← カスタムフック（API呼び出し・状態管理）
  lib/          ← ユーティリティ・型定義・APIクライアント
```

---

## Testing Strategy

### バックエンド（Go）

```go
// テーブル駆動テストの標準形
func TestXxx(t *testing.T) {
    tests := []struct {
        name    string
        input   InputType
        want    OutputType
        wantErr bool
    }{
        {name: "正常系: ...", input: ..., want: ...},
        {name: "異常系: ...", input: ..., wantErr: true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ...
        })
    }
}
```

### フロントエンド（React + Vitest）

```tsx
// コンポーネントテストの標準形
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

test('ボタンクリックでカウントが増える', async () => {
  render(<Counter />)
  await userEvent.click(screen.getByRole('button', { name: /increment/i }))
  expect(screen.getByText('1')).toBeInTheDocument()
})
```

---

## Harness Engineering（ハーネスエンジニアリング）

AI が生成したコードの品質を自動検証するためのルールです。

1. **生成 → テスト実行 → 修正** のサイクルを必ず守る
2. CI が失敗したらコードより先にテストを読む
3. ハーネステスト（統合テスト）で外部依存をモック化する
4. フレイキーテスト（不安定なテスト）は即座に修正する

---

## Environment Variables

| 変数名 | 説明 | 例 |
|--------|------|----|
| `PORT` | バックエンドポート | `8080` |
| `DATABASE_URL` | DB 接続文字列 | `postgres://...` |
| `FRONTEND_ORIGIN` | CORS 許可オリジン | `http://localhost:5173` |
| `GCP_PROJECT_ID` | Google Cloud プロジェクト ID | `my-project` |

---

## API Conventions

- **パス**: `/api/v1/{resource}` 形式
- **レスポンス**: 常に JSON、`Content-Type: application/json`
- **エラー**: `{"error": "message", "code": "ERROR_CODE"}` 形式
- **認証**: `Authorization: Bearer {token}` ヘッダー

---

## Commit & PR Rules

- ブランチ名: `feat/xxx`, `fix/xxx`, `test/xxx`
- PR は 400 行以内を目標
- レビュー前に `make test && make lint` を通す
