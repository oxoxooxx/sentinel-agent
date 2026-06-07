# ---- Build Stage ----
# 使用官方 Go image 編譯，確保 CGO 可用（sqlite3 需要）
FROM golang:1.22-alpine AS builder

# 安裝 C 編譯工具（go-sqlite3 需要 CGO）
RUN apk add --no-cache gcc musl-dev

WORKDIR /build

# 先複製 go.mod / go.sum 以利用 Docker layer cache
COPY go.mod go.sum ./
RUN go mod download

# 複製所有原始碼
COPY . .

# 靜態連結編譯，最小化 binary
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-w -s -extldflags '-static'" \
    -o /sentinel \
    ./cmd/sentinel

# ---- Final Stage ----
# 使用 Alpine 最小化 image（支援 musl libc）
FROM alpine:3.19

# 安裝 ca-certificates（HTTPS 通知需要）與 tzdata（時區支援）
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S sentinel && \
    adduser -S sentinel -G sentinel

WORKDIR /app

# 從 build stage 複製編譯好的 binary
COPY --from=builder /sentinel /app/sentinel

# 建立資料目錄，設定正確的擁有者
RUN mkdir -p /app/data && chown -R sentinel:sentinel /app

# 切換到非 root 用戶（最小權限原則）
USER sentinel

# 預設設定檔路徑（可透過 volume mount 覆寫）
ENV SENTINEL_CONFIG=/app/config.yaml

# UDP syslog 514、HTTP API 8080
EXPOSE 514/udp 8080/tcp

ENTRYPOINT ["/app/sentinel"]
