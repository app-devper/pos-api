# POS API

Point of Sale (POS) REST API — ระบบขายหน้าร้าน รองรับหลายสาขา พร้อมฟีเจอร์ร้านยา

## Features

### Core POS
- **Products** — CRUD, units, prices (multi-tier), stock management, lot tracking, expiry notification
- **Orders** — POS checkout, split payment, bill-level discount, stock deduction
- **Categories** — custom product categories
- **Customers** — CRUD, customer types (General/Wholesaler/Regular)
- **Suppliers** — contact management

### Multi-Branch
- **Branches** — CRUD, branch-scoped data
- **Employees** — linked to UM API, role-based (ADMIN/MANAGER/CASHIER)
- **Stock Transfers** — transfer stock between branches with approve/reject workflow

### Inventory Documents
- **Receives (GR)** — goods receiving with lot creation and stock import

### Reports & Documents (PDF/Excel)
- **Sales Report** — Excel export
- **Stock Report** — Excel export
- **Barcode Labels** — batch barcode/price tag PDF generation
- **PromptPay QR** — EMVCo payload generation + PDF

### Pharmacy (ร้านยา)
- **Drug Info** — drug metadata on products (generic name, type, dosage, contraindications, etc.)
- **Patients** — patient profiles with drug allergy records
- **Allergy Check** — verify products against patient allergies before checkout
- **KHY.9** — drug purchase record (บัญชีการซื้อยา)
- **KHY.10** — specially controlled drug sales record (บัญชีการขายยาควบคุมพิเศษ)
- **KHY.11** — dangerous drug sales record (บัญชีการขายยาอันตราย)
- **KHY.12** — prescription drug sales record (บัญชีการขายยาตามใบสั่งของผู้ประกอบวิชาชีพฯ)
- **KHY.13** — FDA-mandated drug sales report (รายงานการขายยาตามที่เลขาธิการ อย. กำหนด)

### Advanced
- **Dashboard** — daily sales summary, daily chart, low-stock detection
- **Promotions** — percentage/fixed discount rules with product/date conditions
- **Customer History** — activity log per customer
- **Settings** — branch-level config (company info, receipt footer, PromptPay ID, show/hide credit)

### Security
- JWT Authentication
- Redis Session Management
- Role-Based Authorization
- Branch-Scoped Data Access

## Technologies

- [Go](https://go.dev) 1.21+
- [Gin](https://github.com/gin-gonic/gin) — HTTP framework
- [MongoDB](https://www.mongodb.com) — primary database
- [Redis](https://redis.io) — session store
- [fpdf](https://github.com/go-pdf/fpdf) — PDF generation
- [excelize](https://github.com/xuri/excelize) — Excel export

## Setup

Create `.env` file:

```env
PORT=8586
MONGO_HOST=localhost:27017
MONGO_POS_DB_NAME=pos_db
REDIS_HOST=localhost:6379
CLIENT_ID=000
SYSTEM=POS
SECRET_KEY=your_secret_key
```

## Run

```bash
go mod download
go run main.go
```

Dev mode with auto-reload:

```bash
nodemon --exec go run main.go --signal SIGTERM
```

## Cloud Run Cost Tips

โปรเจกต์นี้รองรับการ deploy บน Cloud Run ได้ดีขึ้นแล้วด้วย default ที่ช่วยลด cost:

- ใช้ `GIN_MODE=release` อัตโนมัติบน Cloud Run
- จำกัด Mongo/Redis pool ต่อ instance เพื่อลด connection overhead
- มี HTTP timeout defaults เพื่อตัด request ที่ค้างนานเกินจำเป็น
- ปิด `AUTO_INIT_DEFAULT_BRANCH` บน Cloud Run โดย default เพื่อลด startup I/O

ตัวอย่าง env ที่แนะนำ:

```env
PORT=8080
GIN_MODE=release
MONGO_MAX_POOL_SIZE=10
MONGO_MIN_POOL_SIZE=0
MONGO_MAX_CONN_IDLE_TIME_SEC=120
MONGO_CONNECT_TIMEOUT_SEC=3
MONGO_SERVER_SELECTION_TIMEOUT_SEC=3
REDIS_POOL_SIZE=10
REDIS_MIN_IDLE_CONNS=0
REDIS_DIAL_TIMEOUT_SEC=3
REDIS_READ_TIMEOUT_SEC=3
REDIS_WRITE_TIMEOUT_SEC=3
REDIS_IDLE_TIMEOUT_SEC=120
HTTP_READ_TIMEOUT_SEC=15
HTTP_READ_HEADER_TIMEOUT_SEC=5
HTTP_WRITE_TIMEOUT_SEC=30
HTTP_IDLE_TIMEOUT_SEC=120
```

## API Base Path

```
/api/pos/v1
```

## Documentation

Business logic, workflow, lifecycle, and API contract docs are available under [docs/business-logic](/Users/admin/ProjectPos/devper-pos/pos-api/docs/business-logic/README.md)
