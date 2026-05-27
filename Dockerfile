FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/my-ai-project .

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata netcat-openbsd

ENV TZ=Asia/Shanghai
ENV GIN_MODE=release

COPY --from=builder /app/bin/my-ai-project /app/my-ai-project

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
  CMD nc -z 127.0.0.1 8080 || exit 1

CMD ["/app/my-ai-project"]
