# syntax=docker/dockerfile:1.7
FROM golang:1.22-alpine AS builder

WORKDIR /src
COPY go.mod ./
RUN go mod download || true

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /out/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && adduser -S app -G app

WORKDIR /app
COPY --from=builder /out/server /app/server

USER app
EXPOSE 8080
ENTRYPOINT ["/app/server"]
