FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY order-service/go.mod order-service/go.sum ./
COPY ap2-payment-generated ../ap2-payment-generated
COPY ap2-order-generated ../ap2-order-generated

RUN go mod download

COPY order-service ./

RUN go build -o order-service ./cmd

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/order-service .

CMD ["./order-service"]
