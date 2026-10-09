# テンプレートから新しい案件を始める

このリポジトリは GitHub のテンプレートリポジトリです。
「Use this template」で作成したリポジトリを、案件用に書き換える手順をチェックリストにまとめます。

所要時間の目安は、ローカルで動かすまでが 15 分ほどです。デプロイの準備は [deployment.md](deployment.md) を参照してください。

## 1. リポジトリを作成してツールを入れる

- [ ] GitHub で「Use this template」→「Create a new repository」を選び、案件のリポジトリを作成する
- [ ] クローンして、ツールを入れる

```bash
git clone https://github.com/<org>/<repo>.git
cd <repo>
mise trust && mise install
cp .env.example .env
```

mise の導入と shims の設定は [development.md](development.md) にあります。

## 2. 名前を一括で書き換える

- [ ] ブランチを切り、`mise run init` を実行する

```bash
git switch -c chore/init-from-template
mise run init -- <org>/<repo> --title "画面に表示する名前"
```

`<org>/<repo>` は GitHub のオーナー名とリポジトリ名です。`--title` を省略すると `<repo>` が使われます。

`mise run init` が書き換えるのは次の 4 つです。Git が管理しているテキストファイルすべてが対象です。

| 対象 | テンプレートの値 | 書き換え後 | 主な場所 |
| ---- | ---------------- | ---------- | -------- |
| Go のモジュールパス | `github.com/your-org/webapp-template` | `github.com/<org>/<repo>` | `backend/go.mod`・`cli/go.mod`・全 import |
| リソース名の接頭辞 | `webapp-template` | `<repo>` | Cloud Run のサービス名とジョブ名（`deploy.yml` の `BACKEND_SERVICE`・`FRONTEND_SERVICE`・`MIGRATE_JOB`、`docs/` のコマンド例）、JWT の issuer（`backend/internal/infrastructure/session.go`）、`frontend/package.json` の `name` |
| CLI のバイナリ名 | `webapp-cli` | `<repo>-cli` | `cli/cmd/webapp-cli/`（ディレクトリも改名）、`.gitignore`、`docs/` |
| 画面に表示する名前 | `Webapp Template` | `--title` の値 | `frontend/index.html`、ログイン画面とダッシュボードの見出し |

- [ ] 差分を確認する

```bash
git status
git diff --stat
```

同じコマンドは 2 回実行できません（`backend/go.mod` が書き換え済みだとエラーで止まります）。
やり直すときは `git restore . && git clean -fd cli/cmd` で戻してから実行してください。

JWT の issuer を変えると、変更前に発行したセッションは無効になります。新規の案件では影響ありません。

## 3. 手で決める値

`mise run init` が書き換えない値です。`webapp` のような一般的な語は機械的に置換できないため、必要なものだけ手で直します。

- [ ] **Artifact Registry のリポジトリ名**（既定 `webapp`）。変える場合は `deploy.yml` の `…/webapp/…`（6 か所）と、`docs/deployment.md` の作成コマンドを直す
- [ ] **リージョン**（既定 `asia-northeast1`）。変える場合は `deploy.yml` の `GCP_REGION` と、`docs/` のコマンド例を直す
- [ ] **ローカルの DB 名とイメージ名**（`webapp`・`webapp_test`・`webapp-backend`・`webapp-frontend`）。ローカルでしか使わないので、変えなくても支障はない
- [ ] **ライセンス**。リポジトリに `LICENSE` がある場合は、案件の利用条件に合わせて置き換えるか、削除する
- [ ] **`README.md`** の冒頭を、案件の説明に書き換える

## 4. 使わない OAuth プロバイダを削除する

テンプレートは Google・GitHub・Microsoft の 3 つを実装しています。
クライアント ID を設定しなければそのプロバイダは無効になりますが、ログイン画面のボタンは残ります。使わないものは削除してください。

1 つのプロバイダ（例: `github`）を削除する手順:

- [ ] `backend/cmd/server/main.go` の `GITHUB_CLIENT_ID` を読む `if` ブロックを削除する
- [ ] `backend/internal/infrastructure/oauth_github.go` を削除し、`oauth_test.go` と `oauth_exchange_test.go` から GitHub のテストケースを削除する
- [ ] `frontend/src/pages/LoginPage.tsx` の `providers` から該当の行を削除し、`LoginPage.test.tsx` を合わせる
- [ ] `.env.example` と `docker-compose.yml` から `GITHUB_*` を削除する
- [ ] `deploy.yml` の `--set-secrets` から `GITHUB_CLIENT_ID`・`GITHUB_CLIENT_SECRET` を、`--set-env-vars` から `GITHUB_REDIRECT_URL` を削除する
- [ ] `docs/deployment.md` の Secret Manager の表から該当の行を削除する

Microsoft を使う場合は、`MICROSOFT_TENANT_ID` に自組織のテナントを設定します。未設定だと初回ログインを拒否します（[deployment.md](deployment.md)）。

## 5. ログインできる人を制限する

既定では、OAuth プロバイダのアカウントを持つ人は誰でもユーザー登録でき、誰も管理者になりません。

- [ ] `.env` に、ログインを許可するメールのドメインと、初期管理者のメールアドレスを設定する

```bash
ALLOWED_EMAIL_DOMAINS=example.co.jp
INITIAL_ADMIN_EMAILS=taro@example.co.jp
```

本番では GitHub の Repository Variables に同じ名前で設定します。挙動の詳細は [deployment.md](deployment.md) の「サインアップの制限」にあります。

## 6. 動作を確認する

- [ ] 依存を入れて、CI と同じ検査を通す（Docker が必要）

```bash
mise run setup
mise run fmt
mise run check
```

`mise run check` の内容は [development.md](development.md) の「CI との対応」にあります。Docker がない環境では、`mise run lint`・`mise run test`・`mise run vuln` を個別に実行してください。

- [ ] OAuth のクライアントを作り、`.env` に ID とシークレットを設定する。リダイレクト URL には `http://localhost:5173/api/v1/auth/<provider>/callback` を登録する（`.env.example` の `*_REDIRECT_URL` と同じ値）
- [ ] 起動して、ログインからユーザー一覧の表示までを確認する

```bash
mise run dev
```

- フロントエンド: <http://localhost:5173>
- ヘルスチェック: <http://localhost:5173/health>（frontend 経由で backend に届く）

## 7. コミットして PR を出す

- [ ] コミットして PR を出す

```bash
git add -A
git commit -m "chore: initialize from webapp-template"
git push -u origin chore/init-from-template
```

- [ ] マージ後、このファイル（`docs/getting-started.md`）と `scripts/init-template.mjs`・`scripts/init-template.test.mjs`、`.mise.toml` の `init` タスクを削除する。`docs/README.md` と `README.md` からのリンクも外す

## 8. デプロイの準備

- [ ] [deployment.md](deployment.md) に沿って、GitHub の Secrets / Variables、Workload Identity Federation、Artifact Registry、Secret Manager、Cloud SQL を準備する
- [ ] OAuth プロバイダに、本番のリダイレクト URL（`${FRONTEND_ORIGIN}/api/v1/auth/<provider>/callback`）を登録する
- [ ] `main` にマージして、`Deploy to Cloud Run` ワークフローが成功することを確認する

## 参考: テンプレートに含めていないもの

Pub/Sub・Cloud Functions・バッチ処理などの非同期基盤は、このテンプレートの対象外です。必要な案件では別のリポジトリとして用意してください。
