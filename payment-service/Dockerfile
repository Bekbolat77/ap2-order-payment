FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY payment-service/go.mod payment-service/go.sum ./
COPY ap2-payment-generated ../ap2-payment-generated

RUN go mod download

COPY payment-service ./

RUN go build -o payment-service ./cmd

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/payment-service .

CMD ["./payment-service"]
