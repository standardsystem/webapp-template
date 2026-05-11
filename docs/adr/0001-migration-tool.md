# ADR 0001: DB マイグレーションツールの選定

- **Status**: Accepted
- **Date**: 2026-05-11
- **Deciders**: kazuyukikato
- **Related**: 改善提案書 v2 §6.4, PR3 (migration 分離 + Cloud Run Job 化)

## Context

現状、`backend/cmd/server/main.go` の起動時に自前 runner
(`backend/internal/infrastructure/db.go::RunMigrations`) が `embed.FS` 経由の
`migrations/*.sql` を順次実行している。これは開発体験を最優先にした構成だが、本番
Cloud Run では以下のリスクがある。

- 複数インスタンスの同時起動で同じ migration が並列実行される
- 排他制御 (`pg_advisory_lock`) や 1 migration = 1 transaction の保証なし
- failure 時の rollback / 履歴 (checksum, applied_at) なし
- migration 失敗時に backend service が起動失敗するため、deploy 単位で
  「migration 完走 → backend release」の順序を保証できない

提案書 v2 §6.4 では、本番起動時 migration を廃止し Cloud Run Job または CI/CD の
deploy 前ステップで実行する方針が示されている。実装選択肢として以下の 2 つが挙がっている。

- **A. OSS migration tool 採用** (`goose` / `golang-migrate`) — 優先的に検討
- **B. 自前 runner を強化** — A が合わない場合のみ

本 ADR は PR3 着手前にこの選択を確定する。

## Decision Drivers

1. **Cloud Run Job として実行しやすいこと** — 単独 CLI binary が望ましい
2. **既存の `001_create_users.sql` からの移行コスト**
3. **dev の自動 migration 体験を維持できること** — 現状の `mise run dev` で
   特別な手順なく動くこと
4. **業界での認知度・ドキュメント量** — テンプレートの利用者が学習しやすいこと
5. **本番安全性** — `pg_advisory_lock` 等の排他制御がデフォルトで効くこと
6. **保守負担の少なさ** — 自前コード量を最小化したい

## Considered Options

### Option A1: golang-migrate/migrate

- 公式 CLI binary (`migrate`) と Docker image (`migrate/migrate`) が提供
- ファイル命名規則: `000001_xxx.up.sql` / `000001_xxx.down.sql` (up/down 分離)
- PostgreSQL ドライバが `pg_advisory_lock` を自動取得
- Go API も提供 (embed して dev で auto-migration 継続可能)
- Cloud Run Job 化が最も簡単 (公式 image をそのまま使える)
- Up/Down が常に分離されるため、rollback が習慣化しやすい

#### 移行作業 (golang-migrate)

- `001_create_users.sql` → `000001_create_users.up.sql` に rename
- `000001_create_users.down.sql` を新規作成 (DROP INDEX / DROP TABLE)
- `backend/internal/infrastructure/db.go::RunMigrations` を削除
- dev 用に `cmd/migrate/main.go` を追加 (golang-migrate を embed, または CLI を mise tool 化)
- `docker-compose.yml` の backend command を `migrate up && go run ./cmd/server` に
  するか、別 service として migrate を分離

### Option A2: pressly/goose

- CLI binary と Go API の両方を提供
- ファイル命名規則: 単一ファイル `001_xxx.sql` 内に `-- +goose Up` / `-- +goose Down` ディレクティブ
- PostgreSQL ドライバが `pg_advisory_lock` を自動取得
- 単一ファイルで up/down がペアで管理されるため、PR でレビューしやすい
- Go migration もサポート (本テンプレートでは不要)
- 公式 Docker image なし → 自前で Dockerfile を書くか golang-migrate より手間

#### 移行作業 (goose)

- `001_create_users.sql` の先頭に `-- +goose Up` を追加
- 末尾に `-- +goose Down` セクションを追加 (DROP INDEX / DROP TABLE)
- `backend/internal/infrastructure/db.go::RunMigrations` を削除
- `cmd/migrate/main.go` を追加 (goose embed)
- Cloud Run Job 用に Dockerfile を新設 (binary を含める)

### Option B: 自前 runner を強化

現状の `RunMigrations` を以下で強化:

- `pg_advisory_lock` 取得 → migration 実行 → unlock
- 各 migration を `BEGIN/COMMIT` でラップ
- `schema_migrations` に `checksum`, `applied_at` 追加
- `ENABLE_AUTO_MIGRATION` env で本番起動時 migration を明示許可制に
- `cmd/migrate` を分離して Cloud Run Job 化

#### メリット

- 依存追加なし
- 既存ファイル形式変更不要

#### デメリット

- 自前コードの保守負担 (排他制御 / checksum / rollback ロジック)
- バグ混入時の影響が大きい (本番 DB を破壊する可能性)
- テンプレート利用者が独自仕様を学習する必要がある

## Decision

**Option A1: golang-migrate/migrate を採用する。**

### 理由

1. **Cloud Run Job 化が最も簡単** — 公式 Docker image (`migrate/migrate`) を Cloud Run
   Job の image として直接使える。マイグレーションファイルは Cloud Storage または
   ConfigMap 相当 (Secret Manager / 同梱 image) でマウントするだけ
2. **業界標準** — Go エコシステムで最も普及しており、ドキュメント・blog 記事・
   StackOverflow の情報量が圧倒的に多い。新規参画者の学習コストが最小
3. **Up/Down 分離が rollback 文化を強制する** — テンプレートとして「rollback も
   必ず書く」を規約化できる
4. **自前 runner 廃止** — `db.go::RunMigrations` を削除でき、本テンプレートの保守
   範囲が狭まる
5. **dev の体験は Go API embed で維持可能** — `migrate.NewWithDatabaseInstance` で
   `embed.FS` から自動適用できるため、dev では起動時 auto-migration を継続できる

### 不採用理由

- **goose**: 単一ファイルの可読性は魅力だが、公式 Docker image がないため Cloud Run
  Job 用 Dockerfile を自前管理することになる。テンプレート保守の観点で不利
- **自前 runner**: 排他制御 / checksum / rollback を自前実装する保守コストが、
  golang-migrate 採用コストを上回る合理的な理由が見当たらない

## Consequences

### Positive

- 本番 deploy で migration を deploy_前ステップ (Cloud Run Job) に分離できる
- 排他制御 / transaction / 履歴管理が「公式実装」として担保される
- `RunMigrations` の自前実装を削除でき、テンプレートの保守表面積が減る
- 利用者が他プロジェクトでも同じ tool を使える可能性が高く、横展開しやすい

### Negative

- 新規依存 (`github.com/golang-migrate/migrate/v4`) が backend に追加される
- 既存の `001_create_users.sql` を up/down 分離形式に変換する必要がある
- ファイル数が倍 (up + down) になる

### Neutral

- dev 環境では引き続き起動時 auto-migration を維持するが、実装は golang-migrate
  embed に置き換える (現状の `RunMigrations` は廃止)
- 本番では `ENABLE_AUTO_MIGRATION=false` ではなく、そもそも backend に migration
  実行コードを残さない方針 (混入事故を物理的に防ぐ)

## Implementation Plan (PR3 で実施)

1. `backend/cmd/migrate/main.go` を追加 (golang-migrate を embed して `up` / `down`
   サブコマンドを提供)
2. 既存 `migrations/001_create_users.sql` を以下に分割:
   - `migrations/000001_create_users.up.sql`
   - `migrations/000001_create_users.down.sql`
3. `backend/internal/infrastructure/db.go::RunMigrations` を削除
4. `backend/cmd/server/main.go` から migration 呼び出しを除去
5. dev 用: `docker-compose.yml` の backend service に `command` を migrate → server
   の順で実行する仕掛けを追加 (または別 service `migrate` を起動)
6. 本番用: `.github/workflows/deploy.yml` に `migrate-job` step を追加
   (deploy-backend より先に実行、失敗時は backend deploy をスキップ)
7. `docs/migration.md` を新規作成 (運用手順, dev/本番の違い, naming 規則, rollback)
8. `mise run db:migrate` タスクを `migrate up` 実行に変更

## Follow-ups

- migration 失敗時の Cloud Run Job alerting (Cloud Logging-based alert) は本 ADR の
  範囲外。security PR (PR5) または別途検討
- 既存 dev DB の互換性: schema は同じなので破壊的変更なし。`schema_migrations`
  テーブルの形式が異なる場合は手動で marker を入れる手順を docs に書く
