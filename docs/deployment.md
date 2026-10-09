# デプロイメントガイド

本テンプレートは GitHub Actions から Google Cloud Run へデプロイすることを前提としています。
ワークフロー定義は [`.github/workflows/deploy.yml`](../.github/workflows/deploy.yml) にあります。

## 構成概要

| サービス | Cloud Run service 名 (デフォルト) | ベースイメージ |
|---|---|---|
| backend (Go API) | `webapp-template-api` | distroless |
| frontend (React SPA) | `webapp-template-web` | `nginxinc/nginx-unprivileged:1.30-alpine-slim` |

`webapp-template-api` / `webapp-template-web` は案件ごとに `deploy.yml` の `env:` で書き換えてください。

## フロントエンドから API への経路

利用者のブラウザは frontend のオリジンだけを開きます。frontend の nginx が `/api/` と `/health` を backend へリバースプロキシします。

```text
ブラウザ ──https──▶ frontend (nginx) ──/api/*, /health──▶ backend (Go API)
                        └─ それ以外は静的ファイル（SPA）
```

- フロントのコード（`frontend/src/lib/api.ts`）は相対パス `/api/v1` を呼びます。ブラウザから見てフロントと API が同一オリジンになるため、セッション Cookie（`SameSite=Lax`）がそのまま送られ、CORS の設定に依存しません。
- カスタムドメインがなくても、Cloud Run の既定の URL（`*.run.app`）のままで動きます。`*.run.app` は Public Suffix のため、frontend と backend を別オリジンのまま使うと Cookie が送られません。
- プロキシ先は frontend の環境変数 `BACKEND_URL` です。`deploy.yml` の `deploy-frontend` が、直前にデプロイした backend の URL を `gcloud run services describe` で取得して渡します。手で設定する値はありません。
- OAuth のコールバックも frontend のオリジンで受けます。各プロバイダの管理画面には `${FRONTEND_ORIGIN}/api/v1/auth/{google|github|microsoft}/callback` を登録してください。
- 開発時は Vite の開発サーバーが同じ役割をします（`frontend/vite.config.ts` の `server.proxy`）。転送先は `API_PROXY_TARGET` で、Docker Compose では `http://backend:8080` を渡しています。

この方式の制約:

- API の通信が frontend のインスタンスを経由するため、その分の遅延と課金が増えます。
- backend のサービスは `--allow-unauthenticated` のまま公開されています。backend の URL を直接開いても、Cookie のオリジンが違うためログイン済みの操作はできませんが、到達はできます。
- nginx は起動時に `BACKEND_URL` のホスト名を名前解決します。backend のサービス名を変えたときは frontend も再デプロイしてください（`deploy.yml` は毎回両方をデプロイします）。

社内向けに IAP や Cloud Armor が必要な案件は、外部 HTTPS ロードバランサーでパスを振り分ける構成（`/api/*` → backend、それ以外 → frontend）へ移行してください。
同一オリジンのままなので、フロントのコードと Cookie の設計は変えずに済みます。その場合は `nginx.conf.template` の `/api/` と `/health` の `location` を削除します。

## GitHub 側の設定

デプロイには **Secrets** と **Variables** の両方を設定する必要があります。

### Repository Secrets

| Secret 名 | 用途 |
|---|---|
| `GCP_PROJECT_ID` | Google Cloud プロジェクト ID |
| `GCP_WORKLOAD_IDENTITY_PROVIDER` | Workload Identity Federation のプロバイダー (例: `projects/123/locations/global/workloadIdentityPools/github/providers/github`) |
| `GCP_SERVICE_ACCOUNT` | デプロイ実行用サービスアカウント (例: `deployer@PROJECT.iam.gserviceaccount.com`) |

### Repository Variables

URL 系の値は機密ではないため Secrets ではなく Variables に置きます。

| Variable 名 | 用途 | 例 |
|---|---|---|
| `FRONTEND_ORIGIN` | 利用者が開く公開 URL（frontend のオリジン）。ログイン後のリダイレクト先と、OAuth redirect URL の組立に使用 | `https://app.example.com` |
| `MICROSOFT_TENANT_ID` | Microsoft でのサインインを許可する Entra テナントの ID かドメイン。Microsoft を使うときだけ設定する | `contoso.onmicrosoft.com` |
| `INITIAL_ADMIN_EMAILS` | 初回ログインで `admin` にするメールアドレス（カンマ区切り） | `taro@example.co.jp,hanako@example.co.jp` |
| `ALLOWED_EMAIL_DOMAINS` | ログインを許可するメールのドメイン（カンマ区切り） | `example.co.jp` |

`FRONTEND_ORIGIN` が未設定だと、`deploy.yml` の `Validate required variables` ステップが失敗します。
カスタムドメインを使わない場合は、frontend の Cloud Run サービスの URL（`https://<サービス名>-<プロジェクト番号>.<リージョン>.run.app`）を設定します。
初回は frontend をまだデプロイしていないため、Cloud Run のコンソールか `gcloud run services describe` で URL を確認してから設定し、もう一度デプロイしてください。

バックエンドは、プロバイダが確認したメールアドレスでだけ新規ユーザーを作成します。
Microsoft Graph が返すメールアドレスはテナントの管理者が任意の値に設定できるため、`MICROSOFT_TENANT_ID` で自組織のテナントに限定したときだけ受け付けます。
未設定のときや `common`・`organizations`・`consumers` を指定したときは、Microsoft での初回ログインを 403 で拒否します。

メールアドレスが一致しても、別のプロバイダで登録済みのユーザーには自動で連携しません。
初回ログインのメールアドレスが既存ユーザーと一致した場合は 409 を返します。
プロバイダが確認済みとするメールアドレスでも、退職や再割り当てで現在の所有者が変わっている場合があるためです。

### サインアップの制限

OAuth プロバイダのアカウントを持つ人は、既定では誰でもユーザー登録できます。
社内向けのアプリでは、最初のデプロイの前に次の 2 つを設定してください。

| 環境変数 | 設定したとき | 未設定のとき |
|---|---|---|
| `ALLOWED_EMAIL_DOMAINS` | メールアドレスのドメインが一致する人だけがログインできる。一致しない人は 403 で拒否し、ユーザーも作成しない | ドメインを制限しない。バックエンドは起動時に警告をログに出す |
| `INITIAL_ADMIN_EMAILS` | 一致するメールアドレスの人は、初回ログインで `admin` になる | 誰も `admin` にならない。バックエンドは起動時に警告をログに出す |

- どちらもカンマ区切りで複数指定でき、大文字と小文字を区別しません。
- 「最初にログインした人が `admin` になる」動作はありません。指定していない人は、最初のログインでも `member` です。
- `ALLOWED_EMAIL_DOMAINS` はドメイン全体を比較します。`example.co.jp` を許可しても `sub.example.co.jp` は許可されません。
- `ALLOWED_EMAIL_DOMAINS` は登録済みのユーザーにも適用します。許可するドメインを後から絞ると、外れたユーザーは次回からログインできません。発行済みのセッションは有効期限（24 時間）まで残ります。
- `INITIAL_ADMIN_EMAILS` は初回ログイン（ユーザー作成）のときだけ見ます。すでに `member` として登録済みの人は、追記しても昇格しません。既存の `admin` が `PUT /api/v1/users/{id}/role` でロールを変更してください（テンプレートにロール変更の画面はありません）。`admin` が 1 人もいない場合は、DB の `users.role` を直接更新します。
- 判定に使うのは、プロバイダが確認済みとするメールアドレスです。GitHub は公開メールではなく、検証済みのプライマリメールを使います。

ローカル開発では `.env` に同じ名前で設定します（`.env.example` を参照）。

## Google Cloud 側の設定

### Workload Identity Federation

GitHub Actions から鍵ファイルなしで認証するために設定します。
公式ドキュメント: <https://github.com/google-github-actions/auth#setting-up-workload-identity-federation>

最小権限のサービスアカウントに以下のロールを付与:

- `roles/run.admin` (Cloud Run のデプロイ)
- `roles/iam.serviceAccountUser` (Cloud Run のランタイム SA を引き受ける)
- `roles/artifactregistry.writer` (イメージ push)
- `roles/secretmanager.secretAccessor` (Secret 参照)
- Cloud SQL を使う場合: `roles/cloudsql.client`

### Artifact Registry

イメージは `${REGION}-docker.pkg.dev/${PROJECT_ID}/webapp/${SERVICE}` にプッシュされます。
事前に `webapp` という名前の Docker リポジトリを作成してください。

```bash
gcloud artifacts repositories create webapp \
  --repository-format=docker \
  --location=asia-northeast1
```

### Secret Manager

backend の起動に必要な以下の Secret を Secret Manager に登録します。
`deploy.yml` の `--set-secrets` がこれらを Cloud Run 環境変数にマウントします。

| Secret Manager 名 | 環境変数 | 説明 |
|---|---|---|
| `database-url` | `DATABASE_URL` | PostgreSQL 接続文字列 |
| `jwt-secret` | `JWT_SECRET` | セッション署名用 (32 文字以上) |
| `google-client-id` | `GOOGLE_CLIENT_ID` | Google OAuth |
| `google-client-secret` | `GOOGLE_CLIENT_SECRET` | Google OAuth |
| `github-client-id` | `GITHUB_CLIENT_ID` | GitHub OAuth |
| `github-client-secret` | `GITHUB_CLIENT_SECRET` | GitHub OAuth |
| `microsoft-client-id` | `MICROSOFT_CLIENT_ID` | Microsoft OAuth |
| `microsoft-client-secret` | `MICROSOFT_CLIENT_SECRET` | Microsoft OAuth |

使わない OAuth プロバイダがあれば `deploy.yml` の `--set-secrets` から該当エントリを外してください
(該当 Secret Manager 名は作成不要)。

### Secret 作成例

```bash
echo -n "postgres://USER:PASS@/DBNAME?host=/cloudsql/PROJECT:REGION:INSTANCE" | \
  gcloud secrets create database-url --data-file=-

openssl rand -base64 48 | gcloud secrets create jwt-secret --data-file=-

echo -n "$GOOGLE_CLIENT_ID" | gcloud secrets create google-client-id --data-file=-
# ... 他の OAuth secret も同様
```

## Cloud SQL 接続

本テンプレートでは PostgreSQL を前提にしています。Cloud Run から Cloud SQL に接続する方式は 2 通りあります。

### 方式 A: Cloud SQL connector (Unix socket) — 推奨

設定が簡単で、追加の VPC 構成が不要です。

1. Cloud Run service に Cloud SQL instance を紐づける
2. `DATABASE_URL` を Unix socket 形式にする

```bash
gcloud run services update webapp-template-api \
  --add-cloudsql-instances=PROJECT:REGION:INSTANCE
```

```text
postgres://USER:PASS@/DBNAME?host=/cloudsql/PROJECT:REGION:INSTANCE
```

### 方式 B: Private IP + Serverless VPC Access

VPC 内の他のリソースと組み合わせる場合や、private IP のみで運用したい場合に選択。

1. Serverless VPC Access コネクタを作成
2. Cloud Run に `--vpc-connector` を指定
3. Cloud SQL は private IP で構成
4. firewall / routing / IAM を別途設計

> 詳細は <https://cloud.google.com/sql/docs/postgres/connect-run> を参照。

## frontend の `${PORT}` と `${BACKEND_URL}`

[`frontend/Dockerfile`](../frontend/Dockerfile) は nginx 公式イメージの template 機能を利用して `${PORT}` と `${BACKEND_URL}` を起動時に置換します。

- `frontend/nginx.conf.template` の `listen ${PORT};` が起動時に envsubst で展開される
- 既定値 `PORT=8080` が `Dockerfile` の `ENV` で設定されており、Cloud Run のデフォルト port と一致
- 別 port で起動したい場合 (例: 開発時) は `docker run -e PORT=3000 -p 3000:3000 webapp-frontend`
- `proxy_pass ${BACKEND_URL};` も同じ仕組みで展開される。`BACKEND_URL` にはパスを付けない（例: `https://webapp-template-api-xxxxx.a.run.app`）
- `NGINX_ENVSUBST_FILTER=^(PORT|BACKEND_URL)$` で envsubst の対象をこの 2 つに限定し、nginx の `$uri` などを誤置換しない
- 手元で確認するときは `docker run -e BACKEND_URL=http://host.docker.internal:8080 -p 8080:8080 webapp-frontend`

このため `deploy.yml` の frontend deploy では `--port` を指定していません (Cloud Run のデフォルト 8080 を使用)。

## デプロイ手順

1. 上記 Secrets / Variables / Workload Identity / Artifact Registry / Secret Manager / Cloud SQL を準備
2. `main` ブランチに push (CI 成功後に自動デプロイ)
3. 手動実行する場合は `Actions` タブから `Deploy to Cloud Run` workflow を再実行

## 既知の制限・今後の改善

- DB マイグレーションは Cloud Run Job（`deploy-migrate`）が backend のデプロイ前に実行します。アプリの起動時には実行しません。ロールバックは手動です。手順は [migration.md](migration.md) を参照してください。
- frontend deploy は `--allow-unauthenticated` で公開しています。社内専用にする場合は IAP 等の追加設計が必要です。
