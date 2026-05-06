FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY notification-service/go.mod notification-service/go.sum ./

RUN go mod download

COPY notification-service ./

RUN go build -o notification-service ./cmd

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/notification-service .

CMD ["./notification-service"]
