# AI エージェント向け指示ファイルの構成

Claude Code / Codex / Cursor などのコーディングエージェントに渡す指示は `AGENTS.md` だけに書きます。
`CLAUDE.md` は置きません。

## ファイル配置

```text
AGENTS.md            リポジトリ全体の指示（コマンド・共通ルール・コミット規約）
backend/AGENTS.md    Go API の規約（レイヤー・テスト・マイグレーション）
frontend/AGENTS.md   React / TypeScript の規約
cli/AGENTS.md        Go CLI の規約
```

## どこに何を書くか

| 内容 | 置き場所 | 読み込まれるタイミング |
| ---- | -------- | ---------------------- |
| どのパッケージでも当てはまる指示 | ルートの `AGENTS.md` | 毎セッションの開始時 |
| 1 つのパッケージだけに当てはまる規約 | そのパッケージの `AGENTS.md` | エージェントがそのディレクトリのファイルを扱うとき |
| 特定の作業でだけ必要な手順 | スキル（例: `.cursor/skills/`） | エージェントが必要と判断したとき |
| 人が読む手順書・設計判断 | `docs/` | 読み込まれない（必要なときにエージェントが開く） |

ルートの `AGENTS.md` は毎回コンテキストに載るため、200 行以内に保ちます。
長くなってきたら、パッケージ固有の内容を各パッケージの `AGENTS.md` へ移してください。

## Claude Code での読み込み

Claude Code は **v2.1.277 以降** が必要です。それより前のバージョンは `AGENTS.md` を読まないため、サポートしません。
`claude --version` で確認し、古い場合は更新してください。

v2.1.277 以降の Claude Code は、`CLAUDE.md` がないリポジトリでは `AGENTS.md` を直接読みます。

- セッション開始時に、作業ディレクトリとその上位にある `AGENTS.md` を読みます。
- サブディレクトリの `AGENTS.md` は、そのディレクトリのファイルを開いたときに読みます。
- 読み込まれたファイルは `/context` の **Memory files** で確認できます。

### CLAUDE.md と CLAUDE.local.md を作らない

作業ディレクトリかその上位に `CLAUDE.md`・`.claude/CLAUDE.md`・`CLAUDE.local.md` のどれかがあると、
既定の設定では Claude Code は `AGENTS.md` を読まなくなります。このリポジトリにはこれらを追加しないでください。

個人用の指示を `CLAUDE.local.md` に書きたい場合は、`/config` の **Project instructions** を
`claude-md-and-agents-md` に変更すると、両方が読み込まれます。この設定はユーザー単位で、リポジトリには保存されません。

### Claude Code 固有の指示

Claude Code にだけ伝えたい指示は `.claude/rules/*.md` に置くと、`AGENTS.md` と一緒に読み込まれます。
現在は `.gitignore` が `.claude/` を除外しているため、共有する場合は `.claude/rules/` を除外対象から外してください。

## ツールごとの読み込み方

パッケージの `AGENTS.md` が読み込まれる条件は、ツールによって異なります。

| ツール | ルートの `AGENTS.md` | パッケージの `AGENTS.md` |
| ------ | -------------------- | ------------------------ |
| Claude Code | 起動時に読む | そのディレクトリのファイルを開いたときに読む |
| Cursor | 読む | そのディレクトリ配下のファイルを扱うときに自動で適用する |
| GitHub Copilot | 読む | 作業対象に最も近いものを優先する |
| Codex | 起動時に読む | リポジトリルートから起動ディレクトリまでの経路上にあるものだけ読む |

### Codex は対象パッケージのディレクトリで起動する

Codex は起動時に、リポジトリルートから起動ディレクトリまでの `AGENTS.md` を一度だけ読みます。
ルートで起動すると `backend/AGENTS.md` などは読み込まれません。
作業対象のパッケージのディレクトリで起動してください。ルートとそのパッケージの両方が読み込まれます。

```bash
cd backend && codex
```

複数のパッケージをまたぐ作業をルートから行う場合は、対象パッケージの `AGENTS.md` を読むようプロンプトで指示してください。

## 編集するときの注意

- 「テストを書く」のような曖昧な指示より、「`mise run test:backend` を通す」のように検証できる形で書きます。
- 指示ファイル同士の矛盾や古い記述は、Claude Code の `/doctor prompt-audit` で点検できます。

## 参考

- [How Claude remembers your project](https://code.claude.com/docs/en/memory)
- [Set up Claude Code in a monorepo or large codebase](https://code.claude.com/docs/en/large-codebases)
- [AGENTS.md](https://agents.md/)
- [Codex: AGENTS.md](https://developers.openai.com/codex/guides/agents-md)
- [Cursor: Rules](https://cursor.com/docs/context/rules)
- [GitHub Copilot: repository custom instructions](https://docs.github.com/en/copilot/how-tos/configure-custom-instructions/add-repository-instructions)
