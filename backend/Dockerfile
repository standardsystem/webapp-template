# ---- Build Stage ----
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Alpine OS パッケージに対する最新 CVE パッチを適用
RUN apk upgrade --no-cache

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /server ./cmd/server

# ---- Runtime Stage ----
# distroless static-debian13: Debian 12 系より CVE 蓄積が少ない。:nonroot タグは USER=nonroot を内包
FROM gcr.io/distroless/static-debian13:nonroot

WORKDIR /

COPY --from=builder /server /server

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/server"]
