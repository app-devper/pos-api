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

### Business Documents
- **Purchase Orders (PO)** — CRUD with auto sequence
- **Delivery Orders (DO)** — CRUD with auto sequence
- **Credit Notes (CN)** — CRUD with stock reversal
- **Billings** — CRUD, group multiple orders
- **Quotations** — CRUD with auto sequence
- **Receives (GR)** — goods receiving with lot creation

### Reports & Documents (PDF/Excel)
- **Receipt PDF** — A4 receipt with configurable footer
- **Tax Invoice PDF** — full tax invoice
- **Sales Report** — PDF and Excel
- **Stock Report** — Excel export
- **Receive Summary** — PDF aggregate report
- **Price Report** — PDF with cost/price for all products
- **Product History** — PDF per product or by date range
- **Customer History** — PDF per customer
- **Barcode Labels** — batch barcode/price tag PDF generation
- **PromptPay QR** — EMVCo payload generation + PDF

### Pharmacy (ร้านยา)
- **Drug Info** — drug metadata on products (generic name, type, dosage, contraindications, etc.)
- **Patients** — patient profiles with drug allergy records
- **Allergy Check** — verify products against patient allergies before dispensing
- **Dispensing Logs** — pharmacist dispensing records per order
- **Drug Labels** — auto-generate drug label stickers (70×35mm)
- **KHY.9** — drug receiving report (บัญชีรับยา)
- **KHY.10** — dangerous drug sales report (บัญชีขายยาอันตราย)
- **KHY.11** — specially controlled drug report (บัญชีขายยาควบคุมพิเศษ)
- **KHY.12** — expired drug report
- **KHY.13** — near-expiry drug report

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

## API Base Path

```
/api/pos/v1
```

## API Documentation

Full OpenAPI 3.0 spec available at [`openapi.yaml`](./openapi.yaml). You can view it with:
- [Swagger Editor](https://editor.swagger.io) — paste or import the file
- [Swagger UI](https://petstore.swagger.io) — use URL to the raw file
- VS Code extension: [OpenAPI (Swagger) Editor](https://marketplace.visualstudio.com/items?itemName=42Crunch.vscode-openapi)
