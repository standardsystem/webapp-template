# DB マイグレーションガイド

本テンプレートは [golang-migrate/migrate](https://github.com/golang-migrate/migrate) を採用しています。
選定経緯は [adr/0001-migration-tool.md](adr/0001-migration-tool.md) を参照。

## 設計の原則

- **アプリ起動時 migration は禁止**。backend は `DATABASE_URL` が指す DB を
  「マイグレーション完了済み」として扱う。
- 開発環境では `migrate` service が `compose up` で 1 度だけ実行される。
- 本番環境では Cloud Run Job として deploy パイプラインの一部で実行される。
- `pg_advisory_lock` による排他制御、各 migration の transaction 化は
  golang-migrate が内部で実施する。

## ファイル命名規則

```text
backend/migrations/<version>_<description>.up.sql
backend/migrations/<version>_<description>.down.sql
```

- `version` は 6 桁ゼロ埋め (例: `000001`, `000002`)。番号は単調増加で衝突禁止
- `description` は `snake_case` で短く (例: `create_users`, `add_role_to_users`)
- `up.sql` と `down.sql` は **常に対** で作成する。down が現実的に書けない
  破壊的変更 (例: ALTER TABLE で型を変えてデータを失う) の場合でも、空 down
  ではなく「rollback できないため新 migration で前進復旧」のコメントを残す

## 新しい migration の追加

1. 次の version 番号を決める (既存の最大 + 1)
2. `backend/migrations/000NNN_xxx.up.sql` と `.down.sql` を作成
3. `mise run db:migrate` で local DB に適用して動作確認
4. `mise run db:migrate:down` で down も成功するか必ず確認
5. 必要に応じて `repository` 層の test (integration) を追加

## 開発環境での実行

`mise run dev` (= `docker compose up`) を実行すると、`migrate` service が
DB 起動を待って 1 度だけ migration を実行し、その後 backend が起動する。

手動で実行する場合:

```bash
mise run db:migrate          # up
mise run db:migrate:down     # down 1 step
mise run db:migrate:version  # 現在の version 表示
```

`DATABASE_URL` 環境変数で接続先を変更できる (デフォルトは local compose の DB)。

## 本番環境での実行

`.github/workflows/deploy.yml` の `deploy-migrate` job が Cloud Run Job
`webapp-template-migrate` として実行する。

実行順序:

1. CI 成功 (`check-ci`)
2. `deploy-migrate`: migrate image を build/push、Cloud Run Job を deploy、`gcloud run jobs execute --wait` で実行
3. `deploy-backend`: migration 成功後にのみ実行
4. `deploy-frontend`: backend deploy 後

`deploy-migrate` が失敗すると `deploy-backend` / `deploy-frontend` は
スキップされ、新 backend がスキーマ未適用の DB で起動する事故を防ぐ。

### 手動で migrate Job を実行する

deploy 経由ではなく単発で migration を流したい場合:

```bash
gcloud run jobs execute webapp-template-migrate \
  --region=asia-northeast1 \
  --wait
```

### Cloud SQL 接続

Cloud Run Job も backend service と同様に Cloud SQL 接続設定が必要。

#### Unix socket 方式

```bash
gcloud run jobs deploy webapp-template-migrate \
  --image=... \
  --region=asia-northeast1 \
  --set-secrets="DATABASE_URL=database-url:latest" \
  --add-cloudsql-instances=PROJECT:REGION:INSTANCE
```

`DATABASE_URL` は `postgres://USER:PASS@/DBNAME?host=/cloudsql/PROJECT:REGION:INSTANCE` 形式。

#### Private IP + VPC

```bash
gcloud run jobs deploy webapp-template-migrate \
  --image=... \
  --region=asia-northeast1 \
  --set-secrets="DATABASE_URL=database-url:latest" \
  --vpc-connector=CONNECTOR_NAME
```

詳細は [deployment.md](deployment.md) の Cloud SQL 接続セクションを参照。

## Rollback

### 開発環境

```bash
mise run db:migrate:down
```

### 本番環境

down は **手動オペレーション**として扱う。CI/CD で自動 down は実行しない。

```bash
# 例: backend image は revision n-1 にロールバック、migration は手動で 1 つ戻す
gcloud run services update-traffic webapp-template-api \
  --to-revisions=PREVIOUS_REVISION=100 \
  --region=asia-northeast1

gcloud run jobs execute webapp-template-migrate \
  --args=down \
  --region=asia-northeast1 \
  --wait
```

複数バージョンを戻す場合は `--args="down 3"` のように引数を渡せる
(`force` も同様)。

## Dirty state からの復旧

migration 実行中に異常終了 (例: Cloud Run Job タイムアウト) すると、
`schema_migrations` テーブルに `dirty=true` が記録され、以降の `up` / `down`
が拒否される。

復旧手順:

1. DB の状態を確認し、どこまで適用されたか判断する
2. 実際に適用された version に強制設定する (下記コマンド)
3. 必要に応じて手動で SQL を補正する
4. 通常の `up` を再開する

```bash
# 例: 0000005 まで成功、0000006 で失敗 → 0000005 に強制
gcloud run jobs execute webapp-template-migrate \
  --args="force 5" \
  --region=asia-northeast1 \
  --wait
```

## 禁止事項

- 既に適用済 (= 本番環境にデプロイ済) の migration ファイルを編集しないこと
  (新しい version で修正 migration を追加する)
- `up.sql` だけ作って `down.sql` を省略しないこと
- アプリ起動時 migration を復活させないこと (起動責任の混入は事故源)
- `force` を pipeline で自動実行しないこと (人間判断必須)

## トラブルシューティング

### `Dirty database version N. Fix and force version`

dirty state。「Dirty state からの復旧」を参照。

### `no migration found for version N`

embed したファイルと DB の `schema_migrations` の version が乖離。
`backend/migrations/` のファイル一覧と `gcloud run jobs execute ... --args=version`
の出力を突き合わせる。

### `database is locked`

別プロセスが `pg_advisory_lock` を保持中。Cloud Run Job が並列起動していないか確認。
`max-retries=0` 設定なので通常は単発実行。
