# ---- Build Stage ----
FROM node:24-alpine AS builder

WORKDIR /app

RUN corepack enable && corepack prepare pnpm@10.33.0 --activate

COPY package.json pnpm-lock.yaml* ./
RUN pnpm install --frozen-lockfile

COPY . .
RUN pnpm build

# ---- Runtime Stage ----
FROM nginx:1.28-alpine

COPY --from=builder /app/dist /usr/share/nginx/html
# nginx 1.19+ の公式 entrypoint が /etc/nginx/templates/*.template を envsubst で展開する
COPY nginx.conf.template /etc/nginx/templates/default.conf.template

# Cloud Run は PORT 環境変数で待受ポートを指定する。
# ローカル docker run でも未指定なら 8080。
ENV PORT=8080
# envsubst で置換する変数を ${PORT} のみに限定（nginx の $uri などを誤置換しないため）
ENV NGINX_ENVSUBST_FILTER=^PORT$

EXPOSE 8080

CMD ["nginx", "-g", "daemon off;"]
