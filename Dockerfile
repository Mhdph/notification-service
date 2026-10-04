FROM golang:1.27.1-alpine3.24 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/notification-server \
    ./cmd/server

FROM alpine:3.24

RUN apk add --no-cache ca-certificates \
    && addgroup -S -g 10001 app \
    && adduser -S -D -H -u 10001 -G app app

COPY --from=builder /out/notification-server /usr/local/bin/notification-server

USER 10001:10001
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/notification-server"]
