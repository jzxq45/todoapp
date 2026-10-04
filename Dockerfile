FROM golang:1.22-alpine AS builder

RUN apk add --no-cache gcc musl-dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o /out/todoapp ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates sqlite-libs \
    && addgroup -S -g 10001 todoapp \
    && adduser -S -D -H -u 10001 -G todoapp todoapp \
    && mkdir -p /data \
    && chown todoapp:todoapp /data
WORKDIR /app
COPY --from=builder --chown=todoapp:todoapp /out/todoapp ./todoapp
ENV SERVER_HOST=0.0.0.0 \
    SERVER_PORT=8080 \
    DB_DRIVER=sqlite \
    DB_DSN=/data/todo.db
VOLUME ["/data"]
EXPOSE 8080
USER todoapp
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://127.0.0.1:8080/healthz || exit 1
ENTRYPOINT ["./todoapp"]
