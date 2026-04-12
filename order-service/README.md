# Order Service

## Overview
Order Service is responsible for creating and managing customer orders.

This service follows **Clean Architecture**:
- **domain**: business entity (`Order`)
- **usecase**: business rules and application logic
- **repository**: database operations
- **transport/http**: REST API handlers
- **cmd/main.go**: manual dependency injection (composition root)

The service communicates with **Payment Service** via **REST only** using a custom `http.Client` with a timeout of **2 seconds**.

---

## Responsibilities
- Create new orders
- Store orders in its own database
- Call Payment Service to authorize payment
- Update order status based on payment result
- Return order details
- Cancel only pending orders

---

## Business Rules
1. `amount` must be greater than `0`
2. Money is stored as `int64`
3. New orders are initially saved with status: `Pending`
4. If payment is authorized → order status becomes `Paid`
5. If payment is declined → order status becomes `Failed`
6. If Payment Service is unavailable → Order Service returns `503 Service Unavailable`
7. Only `Pending` orders can be cancelled
8. `Paid` orders cannot be cancelled

---

## Endpoints

### 1. Create Order
**POST** `/orders`

### Request
```json
{
  "customer_id": "cust-1",
  "item_name": "Laptop",
  "amount": 50000
}
```

### Success Response
```json
{
  "id": "generated-order-id",
  "customer_id": "cust-1",
  "item_name": "Laptop",
  "amount": 50000,
  "status": "Paid",
  "created_at": "2026-04-01T12:00:00Z"
}
```

---

### 2. Get Order By ID
**GET** `/orders/{id}`

### Example
`GET /orders/abc123`

---

### 3. Cancel Order
**PATCH** `/orders/{id}/cancel`

### Rule
- Only orders with status `Pending` can be cancelled
- `Paid` orders cannot be cancelled

---

## Environment Variables

- `PORT` → default: `8081`
- `ORDER_DB_DSN` → default: `root:root@tcp(localhost:3306)/order_db?parseTime=true`
- `PAYMENT_SERVICE_URL` → default: `http://localhost:8082`

---

## Database
This service uses its **own database**.

### Table
- `orders`

### Migration
See:
- `migrations/001_create_orders.sql`

---

## Run

### 1. Create database
Example in MySQL:
```sql
CREATE DATABASE order_db;
```

### 2. Run migration
```sql
USE order_db;

CREATE TABLE IF NOT EXISTS orders (
    id VARCHAR(64) PRIMARY KEY,
    customer_id VARCHAR(255) NOT NULL,
    item_name VARCHAR(255) NOT NULL,
    amount BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
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

### Create order
```bash
curl -X POST http://localhost:8081/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id":"cust-1",
    "item_name":"Laptop",
    "amount":50000
  }'
```

### Get order
```bash
curl http://localhost:8081/orders/{id}
```

### Cancel order
```bash
curl -X PATCH http://localhost:8081/orders/{id}/cancel
```

---

## Architecture Notes
- Order Service owns its own data and database
- No shared code with Payment Service
- Order Service does not access Payment database directly
- Inter-service communication is done only through REST
- Handlers are thin
- Business rules are inside the use case layer
- Database logic is inside the repository layer
- Dependencies are wired manually in `main.go`