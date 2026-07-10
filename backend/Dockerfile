FROM postgres:18-alpine AS pg-client

FROM golang:1.26.5-alpine

WORKDIR /app

# Install build dependencies and runtime libraries required by the PostgreSQL
# client copied below.
RUN apk add --no-cache \
    git \
    libpq \
    zstd-libs \
    lz4-libs \
    krb5-libs \
    libldap \
    libedit

# Keep database backup and restore self-contained in this development image.
COPY --from=pg-client /usr/local/bin/pg_dump /usr/local/bin/pg_dump
COPY --from=pg-client /usr/local/bin/psql /usr/local/bin/psql
COPY --from=pg-client /usr/local/lib/libpq.so.5* /usr/local/lib/

# 复制go.mod和go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN VERSION_VALUE="$(./scripts/resolve-version.sh)" && \
    go build -ldflags="-s -w -X main.Version=${VERSION_VALUE}" -o main ./cmd/server/

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./main"]
