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
MONGO_HOST=mongodb://localhost:27017
MONGO_POS_DB_NAME=pos_db
REDIS_HOST=redis://localhost:6379/0
CLIENT_ID=000
SYSTEM=POS
SECRET_KEY=your_secret_key
```

Important notes:

- `MONGO_HOST` must be a MongoDB connection string, for example `mongodb://localhost:27017`
- `REDIS_HOST` must be a Redis URL that `redis.ParseURL` can read, for example `redis://localhost:6379/0`
- Startup now fails fast when required env vars or database/session dependencies are not ready
- Auth middleware now rejects requests when auth config is missing instead of accepting empty values
- Branch fallback to `HQ` happens only when the employee record is genuinely missing, not when the employee lookup fails because of an infrastructure error

## Run

```bash
go mod download
go run main.go
```

Dev mode with auto-reload:

```bash
nodemon --exec go run main.go --signal SIGTERM
```

## Deploy

Deploy ทำอัตโนมัติ ทุก push ที่เข้า branch `main` จะ build และขึ้น Cloud Run
service `pos-dev-api` (`asia-southeast1`) ให้เอง

**build config ไม่ได้อยู่ใน repo นี้** — มันเก็บเป็น inline config บนตัว trigger
เอง แก้ pipeline ต้องไปแก้ที่ trigger ใน Google Cloud ไม่ใช่ที่นี่

```
trigger   deploy-pos-api
project   devperpos
region    global          ← ไม่ใช่ asia-southeast1 ที่ service อยู่
branch    ^main$
```

ดูของจริง:

```bash
gcloud builds triggers describe deploy-pos-api --project=devperpos --region=global
```

ขั้นตอนที่ trigger ทำ:

1. build ด้วย Google Cloud buildpacks (`gcr.io/buildpacks/builder:v1`) — เป็น
   Go buildpack ที่ compile จาก source ตรง ๆ ไม่ได้ใช้ Dockerfile
2. push ขึ้น `asia.gcr.io/devperpos/pos-api/pos-dev-api:$COMMIT_SHA`
3. `gcloud run services update` ด้วย image ของ commit นั้น

### ข้อควรรู้

- **ไม่มีด่าน test ใน pipeline** — deploy ออกไม่ว่าเทสต์จะผ่านหรือไม่ ด่านเดียวที่
  มีคือ workflow `check` ที่รันตอนเปิด PR เพราะฉะนั้นอย่า push เข้า `main` ตรง ๆ
  ให้ผ่าน PR เสมอตาม git flow
- **pipeline ไม่ตั้ง env var ให้** ทุกค่าอยู่บน Cloud Run service และอยู่ข้ามการ
  deploy เพราะ `services update` แก้เฉพาะสิ่งที่ระบุ ตั้งค่าใหม่ด้วย:

  ```bash
  gcloud run services update pos-dev-api \
    --project=devperpos --region=asia-southeast1 \
    --update-env-vars='^##^KEY=value'
  ```

  (`^##^` เปลี่ยน separator เพราะค่าที่มี comma อย่าง `CORS_ALLOWED_ORIGINS`
  จะโดน gcloud ตัดเป็นหลาย env var)
- `CORS_ALLOWED_ORIGINS` ตั้งไว้บน service แล้ว ต้องอัปเดตเองเมื่อเพิ่มหรือเปลี่ยน
  host ของ POS web ถ้าไม่ได้ตั้ง service จะ fallback เป็น `*` พร้อม log warning
- secret อยู่ใน Secret Manager แล้ว service อ้างถึงด้วย `--set-secrets` ไม่ใช่
  เก็บเป็น env var ธรรมดา:

  | Env | Secret |
  |---|---|
  | `MONGO_HOST` | `pos-api-mongo-host` |
  | `REDIS_HOST` | `pos-api-redis-host` |
  | `SECRET_KEY` | `pos-api-secret-key` |

  service account `1056670356976-compute@developer.gserviceaccount.com` มี role
  `roles/secretmanager.secretAccessor` บนทั้งสามตัว หมุนค่าใหม่ด้วยการเพิ่ม
  version แล้ว service จะหยิบไปเองเพราะอ้าง `:latest`:

  ```bash
  printf '%s' 'ค่าใหม่' | gcloud secrets versions add pos-api-secret-key --project=devperpos --data-file=-
  ```

  env var ที่เหลือบน service (`MONGO_POS_DB_NAME`, `CLIENT_ID`, `SYSTEM`,
  `GIN_MODE`, `CORS_ALLOWED_ORIGINS`) ไม่ใช่ความลับ เก็บเป็น plain ต่อไปได้

### Deploy ด้วยมือ

```bash
gcloud builds triggers run deploy-pos-api --project=devperpos --region=global --branch=main
```

## Health Check

| Path | ใช้เมื่อ |
|---|---|
| `GET /health` | ยิงตรงที่ Cloud Run service |
| `GET /api/pos/health` | ยิงผ่าน Firebase gateway (`https://api.devper.app`) เพราะ `/health` ของ gateway ถูก map ไปที่ UM service |

ทั้งสอง path ไม่ต้อง auth และไม่แตะ database — ใช้เป็น liveness signal ล้วน ๆ ส่วน dependency ถูกตรวจด้วย `Ping` ตอน startup อยู่แล้ว

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

## Runtime Hardening

การเปลี่ยนแปลงล่าสุดฝั่ง runtime และ auth มีผลดังนี้:

- server จะไม่ start ต่อถ้า `SECRET_KEY`, `CLIENT_ID`, `SYSTEM`, `MONGO_HOST`, `MONGO_POS_DB_NAME`, หรือ `REDIS_HOST` หาย
- startup จะตรวจ MongoDB และ Redis ด้วย `Ping` ก่อนรับ traffic เพื่อลดกรณี instance ดูเหมือนพร้อมแต่พังที่ request แรก
- ถ้า auth config หายระหว่าง runtime middleware จะตอบ `500` แทนการตรวจ token ด้วยค่า config ว่าง
- branch context จะ fallback ไป `HQ` เฉพาะกรณีไม่พบ employee จริงเท่านั้น ถ้า lookup employee ล้มจาก system error request จะถูกปฏิเสธ
- เพิ่ม regression tests ครอบ startup config validation, auth config validation, และ branch fallback behavior

## API Base Path

```
/api/pos/v1
```

## Documentation

Business logic, workflow, lifecycle, and API contract docs are available under [docs/business-logic](/Users/admin/ProjectPos/devper-pos/pos-api/docs/business-logic/README.md)
