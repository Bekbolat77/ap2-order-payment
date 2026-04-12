# Payment Service

## Overview
Payment Service is responsible for authorizing or declining payments for orders.

This service follows **Clean Architecture**:
- **domain**: business entity (`Payment`)
- **usecase**: business rules and application logic
- **repository**: database operations
- **transport/http**: REST API handlers
- **cmd/main.go**: manual dependency injection (composition root)

This service has its **own database** and does not share storage with any other service.

---

## Responsibilities
- Receive payment requests from Order Service
- Apply payment authorization rules
- Save payment records in its own database
- Return payment result by `order_id`

---

## Business Rules
1. `amount` must be greater than `0`
2. Money is stored as `int64`
3. If `amount <= 100000` → payment status is `Authorized`
4. If `amount > 100000` → payment status is `Declined`
5. A `transaction_id` is generated for every payment

---

## Endpoints

### 1. Create Payment
**POST** `/payments`

### Request
```json
{
  "order_id": "order-123",
  "amount": 50000
}
```

### Success Response
```json
{
  "id": "generated-payment-id",
  "order_id": "order-123",
  "transaction_id": "generated-transaction-id",
  "amount": 50000,
  "status": "Authorized"
}
```

---

### 2. Get Payment By Order ID
**GET** `/payments/{order_id}`

### Example
`GET /payments/order-123`

---

## Environment Variables

- `PORT` → default: `8082`
- `PAYMENT_DB_DSN` → default: `root:root@tcp(localhost:3306)/payment_db?parseTime=true`

---

## Database
This service uses its **own database**.

### Table
- `payments`

### Migration
See:
- `migrations/001_create_payments.sql`

---

## Run

### 1. Create database
Example in MySQL:
```sql
CREATE DATABASE payment_db;
```

### 2. Run migration
```sql
USE payment_db;

CREATE TABLE IF NOT EXISTS payments (
    id VARCHAR(64) PRIMARY KEY,
    order_id VARCHAR(64) NOT NULL,
    transaction_id VARCHAR(64) NOT NULL,
    amount BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL
);
```

### 3. Install dependencies
```bash
go mod tidy
```

### 4. Run service
```bash
go run ./cmd
```

---

## Example cURL

### Create payment
```bash
curl -X POST http://localhost:8082/payments \
  -H "Content-Type: application/json" \
  -d '{
    "order_id":"order-123",
    "amount":50000
  }'
```

### Get payment by order_id
```bash
curl http://localhost:8082/payments/order-123
```

---

## Architecture Notes
- Payment Service owns its own data and database
- No shared code with Order Service
- No direct database access from other services
- Communication with other services is done only through REST
- Handlers are thin
- Business rules are inside the use case layer
- Database logic is inside the repository layer
- Dependencies are wired manually in `main.go`