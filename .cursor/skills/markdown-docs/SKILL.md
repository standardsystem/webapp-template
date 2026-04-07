---
name: markdown-docs
description: >-
  .md を新規作成または大きく編集するとき。.markdownlint-cli2.jsonc に沿い、
  作業後に mise の markdown タスクで検証する。
---

# Markdown ドキュメント

## いつ使うか

- `docs/` またはルートの `*.md` を新規・大幅更新するとき
- README / AGENTS / CLAUDE まわりを触るとき

## 手順

1. `.cursor/rules/markdown.mdc` のチェックリストに沿って書く。
2. 人向けの長文は `docs/` に置き、ルートは概要とリンクに留める（`AGENTS.md` / `CLAUDE.md` と矛盾させない）。
3. 完了前にリポジトリルートで `mise run fmt:markdown` と `mise run lint:markdown` を実行し、修正可能な違反は直す。
4. `lint:markdown` が通らない場合は、設定で緩めているルール（例: `MD013`）以外は本文側を直す。

## 参照

- `.markdownlint-cli2.jsonc`
- リポジトリルートの `docs/README.md`
