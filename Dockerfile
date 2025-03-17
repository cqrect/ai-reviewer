FROM golang:1.23 AS builder
WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-s -w' -o review main.go

FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache \
    ca-certificates \
    tzdata && \
    update-ca-certificates
COPY --from=builder /build/review /app/review

ENTRYPOINT ["/app/review"]
