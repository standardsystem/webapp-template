# frontend — React + Vite + TypeScript

`frontend/` で作業するときの規約です。リポジトリ共通の指示はルートの `AGENTS.md` にあります。

## 構成

```text
src/components/   UI コンポーネント（ロジックを持たせない）
src/pages/        ルーティング単位の画面
src/hooks/        カスタムフック（API 呼び出し・状態管理）
src/contexts/     React Context
src/lib/          ユーティリティ・型定義・API クライアント
src/test/         テストのセットアップ
```

## コーディング

- `tsconfig.json` の `strict: true` を維持する。
- `any` を使わない。型が不明な値は `unknown` で受けて絞り込む。
- コンポーネントは関数コンポーネントで書く。クラスコンポーネントは使わない。
- API 呼び出しや状態管理のロジックはカスタムフックに分離する。
- パッケージマネージャは pnpm。`npm` / `yarn` でロックファイルを作らない。

## テスト

- Vitest + Testing Library で書き、実装と同じディレクトリに `*.test.tsx` で置く。
- 要素は `getByRole` などユーザーから見える手がかりで取得する。
- カバレッジの目標は 70% 以上。

```bash
mise run test:frontend   # vitest run
mise run lint:frontend   # eslint + tsc --noEmit
```
