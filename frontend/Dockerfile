# ---- Build Stage ----
# trixie-slim (Debian 13) は alpine 系より OS パッケージ更新が頻繁で CVE 数が少ない。
# Vite の出力は静的ファイルなので libc 差 (musl→glibc) は成果物に影響しない。
FROM node:24-trixie-slim AS builder

WORKDIR /app

# Debian パッケージに対する最新 CVE パッチを適用
RUN apt-get update \
  && apt-get upgrade -y \
  && rm -rf /var/lib/apt/lists/* \
  && corepack enable \
  && corepack prepare pnpm@10.33.0 --activate

COPY package.json pnpm-lock.yaml* ./
RUN pnpm install --frozen-lockfile

COPY . .
RUN pnpm build

# ---- Runtime Stage ----
# nginx-unprivileged の slim variant:
#   - 非 root (uid 101 / nginx) で起動するため Cloud Run の最小権限要件に合致
#   - slim は brotli/geoip/image-filter/perl/xslt 等の不要モジュールを含まず CVE 表面が小さい
#   - 1.30 系は現行 stable
FROM nginxinc/nginx-unprivileged:1.30-alpine-slim

# Alpine OS パッケージに対する最新 CVE パッチを適用（root 権限が必要）
USER root
RUN apk upgrade --no-cache
USER nginx

COPY --from=builder --chown=nginx:nginx /app/dist /usr/share/nginx/html
# nginx 1.19+ の公式 entrypoint が /etc/nginx/templates/*.template を envsubst で展開する
COPY --chown=nginx:nginx nginx.conf.template /etc/nginx/templates/default.conf.template

# Cloud Run は PORT 環境変数で待受ポートを指定する。
# nginx-unprivileged のデフォルト listen も 8080 で一致。
ENV PORT=8080
# envsubst で置換する変数を ${PORT} のみに限定（nginx の $uri などを誤置換しないため）
ENV NGINX_ENVSUBST_FILTER=^PORT$

EXPOSE 8080

CMD ["nginx", "-g", "daemon off;"]
