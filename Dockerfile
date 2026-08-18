# ---- Build stage ----
FROM golang:1.25-alpine AS builder
WORKDIR /src

# Cache module downloads.
COPY go.mod go.sum ./
RUN go mod download

# Copy source (the frontend is pre-built and embedded in internal/static/dist).
COPY . .

# Pure-Go binary: CGO disabled, static assets embedded, cross-compile safe.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/chinasclmdav .

# ---- Runtime stage ----
FROM alpine:3.20
# 固定 UID/GID=1000，宿主机挂载目录 chown -R 1000:1000 即可
RUN addgroup -S -g 1000 scldav \
    && adduser -S -D -H -u 1000 -G scldav scldav \
    && apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /out/chinasclmdav /app/chinasclmdav

# Default data volume (user files + SQLite db).
ENV CHINASCLMDAV_DATA=/data \
    CHINASCLMDAV_LISTEN=:8080 \
    CHINASCLMDAV_PUBLIC_URL=http://localhost:8080

VOLUME ["/data"]
EXPOSE 8080

USER scldav
ENTRYPOINT ["/app/chinasclmdav"]
CMD ["-listen", ":8080", "-data", "/data"]
