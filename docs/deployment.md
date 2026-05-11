# デプロイメントガイド

本テンプレートは GitHub Actions から Google Cloud Run へデプロイすることを前提としています。
ワークフロー定義は [`.github/workflows/deploy.yml`](../.github/workflows/deploy.yml) にあります。

## 構成概要

| サービス | Cloud Run service 名 (デフォルト) | ベースイメージ |
|---|---|---|
| backend (Go API) | `webapp-template-api` | distroless |
| frontend (React SPA) | `webapp-template-web` | `nginx:1.28-alpine` |

`webapp-template-api` / `webapp-template-web` は案件ごとに `deploy.yml` の `env:` で書き換えてください。

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
| `FRONTEND_ORIGIN` | フロントエンド公開 URL。CORS / Cookie 検証で使用 | `https://app.example.com` |
| `BACKEND_ORIGIN` | バックエンド公開 URL。OAuth redirect URL の組立に使用 | `https://api.example.com` |

`deploy.yml` の `Validate required variables` ステップで未設定時は失敗します。

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

## frontend の `${PORT}` 対応

[`frontend/Dockerfile`](../frontend/Dockerfile) は nginx 公式イメージの template 機能を利用して `${PORT}` を起動時に置換します。

- `frontend/nginx.conf.template` の `listen ${PORT};` が起動時に envsubst で展開される
- 既定値 `PORT=8080` が `Dockerfile` の `ENV` で設定されており、Cloud Run のデフォルト port と一致
- 別 port で起動したい場合 (例: 開発時) は `docker run -e PORT=3000 -p 3000:3000 webapp-frontend`
- `NGINX_ENVSUBST_FILTER=^PORT$` で envsubst の対象を `${PORT}` のみに限定し、nginx の `$uri` などを誤置換しない

このため `deploy.yml` の frontend deploy では `--port` を指定していません (Cloud Run のデフォルト 8080 を使用)。

## デプロイ手順

1. 上記 Secrets / Variables / Workload Identity / Artifact Registry / Secret Manager / Cloud SQL を準備
2. `main` ブランチに push (CI 成功後に自動デプロイ)
3. 手動実行する場合は `Actions` タブから `Deploy to Cloud Run` workflow を再実行

## ローカルからの手動デプロイ

開発時の確認やトラブルシュート用に、ローカルから直接 deploy したい場合は `Makefile` の `make deploy` ターゲットを利用できます (個別環境設定が必要)。

## 既知の制限・今後の改善

- 本テンプレートは現状 **アプリ起動時に DB マイグレーションを自動実行** します。本番運用では PR3 で予定している Cloud Run Job 化への移行が必須です。詳細は提案書 v2 を参照。
- frontend deploy は `--allow-unauthenticated` で公開しています。社内専用にする場合は IAP 等の追加設計が必要です。
