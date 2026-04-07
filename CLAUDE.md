# CLAUDE.md — AI 協働開発ガイド（要約）

> **完全版:** [docs/CLAUDE.md](docs/CLAUDE.md)

---

## 一言概要

Go API + Go CLI + React の **モノレポ** テンプレート。ツールは **mise**、人向け Docs は **`docs/`**。

---

## 開発の要点

- API の依存方向: `handler → usecase → domain ← repository`
- `mise run setup` → `mise run dev` / `mise run check`
- Markdown は `.markdownlint-cli2.jsonc` に合わせる（行長 `MD013` は無効）

---

## 詳細リンク

- [docs/CLAUDE.md](docs/CLAUDE.md) — ディレクトリ構造・コマンド一覧・ルール全文
- [docs/development.md](docs/development.md) — セットアップと pre-commit
- [docs/monorepo.md](docs/monorepo.md) — Web + CLI 構成
