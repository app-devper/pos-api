# POS API — Requirements Summary

ระบบ Point of Sale (POS) REST API รองรับหลายสาขา พร้อมโมดูลร้านยาครบวงจร

---

## 1. ภาพรวมระบบ

| รายการ | รายละเอียด |
|--------|-----------|
| Base URL | `/api/pos/v1` |
| Protocol | REST / JSON |
| Auth | JWT Bearer Token + Redis Session |
| Database | MongoDB |
| Language | Go 1.21+ |
| Framework | Gin |

---

## 2. Authentication & Authorization

- **JWT Bearer Token** — ทุก endpoint ต้องมี `Authorization: Bearer <token>` (ยกเว้นที่ระบุ `security: []`)
- **Redis Session** — middleware `RequireSession` ตรวจสอบ session จาก Redis และฝัง `UserId`, `BranchId` ใน context
- **Role-Based Access**

| Role | สิทธิ์ |
|------|--------|
| `ADMIN` | CRUD ทุกอย่าง รวม delete / ลบ order / approve transfer |
| `MANAGER` | อ่านข้อมูล + บางส่วนของ write |
| `CASHIER` | สร้าง order, อ่านข้อมูลสินค้า/ลูกค้า |

- **Branch-Scoped** — ข้อมูลทุก collection ผูกกับ `branchId` จาก session token โดยอัตโนมัติ

---

## 3. Feature Modules

### 3.1 Products (สินค้า)

**Endpoints**

| Method | Path | คำอธิบาย | Role |
|--------|------|-----------|------|
| GET | `/products` | ดึงสินค้าทั้งหมด (พร้อม units/prices/stocks) | All |
| POST | `/products` | สร้างสินค้า | ADMIN |
| GET | `/products/{productId}` | ดึงสินค้า by ID | All |
| PUT | `/products/{productId}` | อัปเดตสินค้า | ADMIN |
| DELETE | `/products/{productId}` | ลบสินค้า | ADMIN |
| POST | `/products/receive` | สร้างสินค้าพร้อมรับของเข้า (ครั้งแรก) | ADMIN |
| GET | `/products/serial-number` | Generate serial number ใหม่ | All |
| GET | `/products/serial-number/{sn}` | ค้นหาสินค้า by serial number | All |
| DELETE | `/products/{productId}/sold-first` | Clear sold-first quantity | ADMIN |

**Business Rules**
- Serial number ต้องไม่ซ้ำ
- การสร้างสินค้าครั้งแรก (`/products/receive`) จะสร้าง unit, price, stock และ product history โดยอัตโนมัติ
- ไม่สามารถลบ default unit (size=1) ได้
- ไม่สามารถลบ default price (CustomerType=GENERAL) ได้

---

### 3.2 Product Units (หน่วยสินค้า)

| Method | Path | คำอธิบาย |
|--------|------|-----------|
| GET | `/products/{productId}/units` | ดึง units ทั้งหมดของสินค้า |
| POST | `/products/units` | สร้าง unit |
| PUT | `/products/units/{unitId}` | อัปเดต unit |
| DELETE | `/products/units/{unitId}` | ลบ unit (ไม่สามารถลบ default ได้) |

---

### 3.3 Product Prices (ราคา)

| Method | Path | คำอธิบาย |
|--------|------|-----------|
| GET | `/products/{productId}/prices` | ดึง prices ทั้งหมด |
| POST | `/products/prices` | สร้าง price tier |
| PUT | `/products/prices/{priceId}` | อัปเดต price |
| DELETE | `/products/prices/{priceId}` | ลบ price (ไม่สามารถลบ GENERAL ได้) |

**Customer Types:** `GENERAL` / `WHOLESALER` / `REGULAR`

---

### 3.4 Product Stocks (สต็อก)

| Method | Path | คำอธิบาย |
|--------|------|-----------|
| GET | `/products/{productId}/stocks` | ดึง stocks ทั้งหมดของสินค้า |
| POST | `/products/stocks` | สร้าง stock lot |
| PUT | `/products/stocks/{stockId}` | อัปเดต stock |
| DELETE | `/products/stocks/{stockId}` | ลบ stock |
| PATCH | `/products/stocks/{stockId}/quantity` | อัปเดตจำนวน |
| PATCH | `/products/stocks/sequence` | จัดลำดับ FIFO sequence |

**Business Rules**
- แต่ละ stock มี sequence สำหรับจัดลำดับการตัดสต็อก (FIFO)
- ทุก stock operation บันทึก product history โดยอัตโนมัติ
- `balance` คำนวณจาก `SUM(quantity)` ของ stocks ทั้งหมดของ product+unit

---

### 3.5 Product Lots & Expiry

| Method | Path | คำอธิบาย |
|--------|------|-----------|
| GET | `/products/lots/expire-notify` | ดึง lot ที่กำลังจะหมดอายุ (public) |

---

### 3.6 Orders (คำสั่งซื้อ / POS)

| Method | Path | คำอธิบาย | Role |
|--------|------|-----------|------|
| POST | `/orders` | สร้าง order (checkout) | All |
| GET | `/orders` | ดึง orders by date range | All |
| GET | `/orders/{orderId}` | ดึง order detail by ID | All |
| DELETE | `/orders/{orderId}` | ลบ order + คืน stock | ADMIN |
| PATCH | `/orders/{orderId}/customer-code` | แนบ customer code กับ order | All |
| DELETE | `/orders/{orderId}/products/{productId}` | ลบ order item (by order+product) | ADMIN |
| GET | `/orders/customers/{customerCode}` | ดึง orders by customer | All |
| GET | `/orders/items` | ดึง order items by date range | All |
| GET | `/orders/items/{itemId}` | ดึง order item by ID | All |
| DELETE | `/orders/items/{itemId}` | ลบ order item | ADMIN |
| GET | `/orders/items/products/{productId}` | ดึง order items by product | ADMIN |
| GET | `/orders/item-details/products/{productId}` | ดึง order item details by product | ADMIN |

**Business Rules**
- สร้าง order = ตัด stock ทันที (ตาม `stocks[]` ที่ระบุใน request)
- รองรับ split payment (หลาย payment type)
- รองรับ discount ระดับ bill
- ลบ order = คืน stock ทุกรายการ
- ทุก order item operation บันทึก product history

---

### 3.7 Categories (หมวดหมู่)

| Method | Path | คำอธิบาย |
|--------|------|-----------|
| GET | `/categories` | ดึงทั้งหมด |
| POST | `/categories` | สร้าง |
| GET | `/categories/{id}` | ดึง by ID |
| PUT | `/categories/{id}` | อัปเดต |
| DELETE | `/categories/{id}` | ลบ |
| PATCH | `/categories/{id}/default` | ตั้งเป็น default category |

---

### 3.8 Customers (ลูกค้า)

| Method | Path | คำอธิบาย |
|--------|------|-----------|
| GET | `/customers` | ดึงทั้งหมด |
| POST | `/customers` | สร้าง (code ต้องไม่ซ้ำ) |
| GET | `/customers/{id}` | ดึง by ID |
| PUT | `/customers/{id}` | อัปเดต |
| DELETE | `/customers/{id}` | ลบ |
| PATCH | `/customers/{id}/status` | อัปเดต status |
| GET | `/customers/code/{code}` | ดึง by customer code |

**Customer Types:** `GENERAL` / `WHOLESALER` / `REGULAR`

---

### 3.9 Suppliers (ผู้จัดจำหน่าย)

| Method | Path | คำอธิบาย |
|--------|------|-----------|
| GET | `/suppliers` | ดึงทั้งหมด |
| POST | `/suppliers` | สร้าง |
| GET | `/suppliers/{id}` | ดึง by ID |
| PUT | `/suppliers/{id}` | อัปเดต |
| DELETE | `/suppliers/{id}` | ลบ |
| GET | `/suppliers/info` | ดึงข้อมูล supplier ของสาขา |
| PUT | `/suppliers/info` | Upsert ข้อมูล supplier ของสาขา |

---

### 3.10 Receives / Goods Receiving (GR)

| Method | Path | คำอธิบาย |
|--------|------|-----------|
| GET | `/receives` | ดึง by date range |
| POST | `/receives` | สร้าง GR |
| GET | `/receives/{id}` | ดึง by ID |
| PUT | `/receives/{id}` | อัปเดต |
| DELETE | `/receives/{id}` | ลบ |
| PATCH | `/receives/{id}/total-cost` | อัปเดต total cost |
| PATCH | `/receives/{id}/items` | อัปเดต items |

---

### 3.11 Branches (สาขา)

| Method | Path | คำอธิบาย | Role |
|--------|------|-----------|------|
| GET | `/branches` | ดึงทั้งหมด | All |
| POST | `/branches` | สร้าง | ADMIN |
| GET | `/branches/{id}` | ดึง by ID | All |
| PUT | `/branches/{id}` | อัปเดต | ADMIN |
| DELETE | `/branches/{id}` | ลบ | ADMIN |
| PATCH | `/branches/{id}/status` | เปิด/ปิด สาขา | ADMIN |

---

### 3.12 Employees (พนักงาน)

| Method | Path | คำอธิบาย | Role |
|--------|------|-----------|------|
| GET | `/employees` | ดึงทั้งหมด | All |
| POST | `/employees` | สร้าง (link กับ UM API) | ADMIN |
| GET | `/employees/{id}` | ดึง by ID | All |
| PUT | `/employees/{id}` | อัปเดต | ADMIN |
| DELETE | `/employees/{id}` | ลบ | ADMIN |
| GET | `/employees/branch/{branchId}` | ดึง by branch | All |

---

### 3.13 Dashboard

| Method | Path | คำอธิบาย | Parameters |
|--------|------|-----------|-----------|
| GET | `/dashboard/summary` | ยอดขาย รายรับ ต้นทุน กำไร | startDate, endDate |
| GET | `/dashboard/daily-chart` | ยอดขายรายวัน (max 31 วัน) | startDate, endDate |
| GET | `/dashboard/low-stock` | สินค้าสต็อกต่ำ | threshold (default 10) |
| GET | `/dashboard/stock-report` | รายงานสต็อกรวม | - |

---

### 3.14 Settings (การตั้งค่าสาขา)

| Method | Path | คำอธิบาย | Role |
|--------|------|-----------|------|
| GET | `/settings` | ดึงการตั้งค่าของสาขา | All |
| PUT | `/settings` | Upsert การตั้งค่า | ADMIN |

**Fields:** companyName, address, phone, taxId, receiptFooter, promptPayId, showCredit

---

### 3.15 Business Documents

#### Purchase Orders (PO)
| Method | Path | Role |
|--------|------|------|
| GET/POST | `/purchase-orders` | ADMIN (POST) |
| GET/PUT/DELETE | `/purchase-orders/{id}` | ADMIN (PUT/DELETE) |

#### Delivery Orders (DO)
| Method | Path | Role |
|--------|------|------|
| GET/POST | `/delivery-orders` | ADMIN (POST) |
| GET/PUT/DELETE | `/delivery-orders/{id}` | ADMIN (PUT/DELETE) |

#### Credit Notes (CN)
- สร้าง credit note = **คืน stock โดยอัตโนมัติ**

| Method | Path | Role |
|--------|------|------|
| GET/POST | `/credit-notes` | ADMIN (POST) |
| GET/PUT/DELETE | `/credit-notes/{id}` | ADMIN (PUT/DELETE) |

#### Billings
| Method | Path | Role |
|--------|------|------|
| GET/POST | `/billings` | ADMIN (POST) |
| GET/PUT/DELETE | `/billings/{id}` | ADMIN (PUT/DELETE) |

#### Quotations
| Method | Path | Role |
|--------|------|------|
| GET/POST | `/quotations` | ADMIN (POST) |
| GET/PUT/DELETE | `/quotations/{id}` | ADMIN (PUT/DELETE) |

**ทุก document มี auto-sequence** (PO-0001, DO-0001, CN-0001, QT-0001, ST-0001)

---

### 3.16 Promotions (โปรโมชั่น)

| Method | Path | คำอธิบาย |
|--------|------|-----------|
| GET | `/promotions` | ดึงทั้งหมด |
| POST | `/promotions` | สร้าง |
| GET | `/promotions/{id}` | ดึง by ID |
| PUT | `/promotions/{id}` | อัปเดต |
| DELETE | `/promotions/{id}` | ลบ |
| POST | `/promotions/apply` | Apply promotion code ให้ order |

**Discount Types:** `PERCENT` / `FIXED`  
**Conditions:** productIds, date range, minimum amount

---

### 3.17 Customer History

| Method | Path | คำอธิบาย |
|--------|------|-----------|
| POST | `/customer-histories` | บันทึก activity |
| GET | `/customer-histories/{customerCode}` | ดึง history by customer code |

---

### 3.18 Stock Transfers (โอนสต็อกระหว่างสาขา)

| Method | Path | คำอธิบาย | Role |
|--------|------|-----------|------|
| GET | `/stock-transfers` | ดึง transfers ของสาขา (from/to) | All |
| POST | `/stock-transfers` | สร้าง transfer request | ADMIN |
| GET | `/stock-transfers/{id}` | ดึง by ID | All |
| PATCH | `/stock-transfers/{id}/approve` | Approve → ตัดสต็อกต้นทาง + เพิ่มสต็อกปลายทาง | ADMIN |
| PATCH | `/stock-transfers/{id}/reject` | Reject | ADMIN |

**Workflow:** `PENDING` → `APPROVED` หรือ `REJECTED`  
- Approve แล้วไม่สามารถย้อนกลับได้ (filter status=PENDING ก่อน update เพื่อป้องกัน race condition)

---

## 4. Pharmacy Module (ร้านยา)

### 4.1 Drug Info (ข้อมูลยา)
ฝังอยู่ใน Product entity

| Field | ประเภท | คำอธิบาย |
|-------|--------|-----------|
| genericName | string | ชื่อสามัญ |
| drugType | string | `DANGEROUS` / `SPECIALLY_CONTROLLED` / อื่นๆ |
| dosageForm | string | รูปแบบยา |
| strength | string | ความแรง |
| indication | string | ข้อบ่งใช้ |
| dosage | string | ขนาดใช้ยา |
| sideEffects | string | ผลข้างเคียง |
| contraindications | string | ข้อห้ามใช้ |
| storageCondition | string | การเก็บรักษา |
| manufacturer | string | ผู้ผลิต |
| registrationNo | string | เลขทะเบียนยา |
| isControlled | bool | ยาควบคุม |

---

### 4.2 Patients (ผู้ป่วย)

| Method | Path | คำอธิบาย |
|--------|------|-----------|
| GET | `/patients` | ดึงทั้งหมด (by branch) |
| POST | `/patients` | สร้าง patient profile |
| GET | `/patients/{id}` | ดึง by ID |
| PUT | `/patients/{id}` | อัปเดต |
| DELETE | `/patients/{id}` | ลบ |
| GET | `/patients/customer/{customerCode}` | ดึง by customer code |
| POST | `/patients/{id}/allergy-check` | ตรวจสอบการแพ้ยา |

**Patient Data:** idCard, dateOfBirth, gender, bloodType, weight, allergies[], chronicDiseases[], currentMedications[], note

**Allergy Check:**
- รับ `productIds[]` → ตรวจสอบ product name/generic name กับ allergy ของ patient
- Return `warnings[]` พร้อม drugName, reaction, severity

---

### 4.3 Dispensing Logs (บันทึกการจ่ายยา)

| Method | Path | คำอธิบาย |
|--------|------|-----------|
| GET | `/dispensing-logs` | ดึงทั้งหมด (by branch) |
| POST | `/dispensing-logs` | บันทึกการจ่ายยา |
| GET | `/dispensing-logs/{id}` | ดึง by ID |
| GET | `/dispensing-logs/patient/{patientId}` | ดึง by patient |

**Fields:** orderId, patientId, items[], pharmacistName, licenseNo, note

---

## 5. Reports & Documents

### 5.1 PDF Reports

| Endpoint | คำอธิบาย | Parameters |
|----------|-----------|-----------|
| `GET /reports/receipt/{orderId}/pdf` | ใบเสร็จ A4 | - |
| `GET /reports/tax-invoice/{orderId}/pdf` | ใบกำกับภาษี | - |
| `GET /reports/sales/pdf` | รายงานยอดขาย | startDate, endDate |
| `GET /reports/product-history/{productId}/pdf` | ประวัติสินค้า | - |
| `GET /reports/product-history/pdf` | ประวัติสินค้า by date range | startDate, endDate |
| `GET /reports/customer-history/{code}/pdf` | ประวัติลูกค้า | - |
| `GET /reports/receives/summary/pdf` | รายงานรับสินค้า | startDate, endDate |
| `GET /reports/prices/pdf` | รายงานราคาสินค้า | category (optional) |
| `GET /reports/price-tags/pdf` | ป้ายราคา (ทุกสินค้า) | category (optional) |
| `POST /reports/barcodes/pdf` | Barcode labels (batch) | productIds[] |
| `GET /reports/drug-label/{logId}/pdf` | สติ๊กเกอร์ยา 70×35mm | - |

### 5.2 Excel Reports

| Endpoint | คำอธิบาย | Parameters |
|----------|-----------|-----------|
| `GET /reports/sales/excel` | รายงานยอดขาย | startDate, endDate |
| `GET /reports/stocks/excel` | รายงานสต็อก | - |

### 5.3 Pharmacy Reports (PDF)

| Endpoint | คำอธิบาย | Parameters |
|----------|-----------|-----------|
| `GET /reports/pharmacy/khy9` | บัญชีรับยา | startDate, endDate |
| `GET /reports/pharmacy/khy10` | บัญชีขายยาอันตราย | startDate, endDate |
| `GET /reports/pharmacy/khy11` | บัญชีขายยาควบคุมพิเศษ | startDate, endDate |
| `GET /reports/pharmacy/khy12` | รายงานยาหมดอายุ | - |
| `GET /reports/pharmacy/khy13` | รายงานยาใกล้หมดอายุ | days (default 90) |

### 5.4 PromptPay QR

Generate EMVCo QR payload ตาม PromptPay ID ที่กำหนดใน branch settings

---

## 6. Product History (การเคลื่อนไหวสินค้า)

บันทึกอัตโนมัติทุกครั้งที่:
- สร้าง/รับสินค้าใหม่ → type `RECEIVE`
- สร้าง order → type `SALE`
- ลบ order → type `RETURN`
- เพิ่ม/ลด stock → type `STOCK_IN` / `STOCK_OUT`
- อัปเดตราคา → type `PRICE_UPDATE`
- อัปเดตข้อมูลสินค้า → type `PRODUCT_UPDATE`

**Fields:** productId, branchId, type, unit, quantity, balance, costPrice, price, description, createdBy, createdDate

---

## 7. Sequence Numbers (เลขที่เอกสาร)

| ประเภทเอกสาร | Constant | ตัวอย่าง |
|-------------|---------|---------|
| Purchase Order | `PURCHASE_ORDER` | PO-0001 |
| Delivery Order | `DELIVERY_ORDER` | DO-0001 |
| Credit Note | `CREDIT_NOTE` | CN-0001 |
| Billing | `BILLING` | BL-0001 |
| Quotation | `QUOTATION` | QT-0001 |
| Stock Transfer | `STOCK_TRANSFER` | ST-0001 |

Sequence เก็บใน `sequences` collection และ increment ด้วย MongoDB `$inc`

---

## 8. MongoDB Collections

| Collection | คำอธิบาย | Index หลัก |
|------------|---------|-----------|
| `products` | สินค้า | `serialNumber` |
| `product_units` | หน่วยสินค้า | `productId` |
| `product_prices` | ราคาสินค้า | `productId` |
| `product_stocks` | สต็อกสินค้า | `branchId+productId`, `productId+unitId` |
| `product_lots` | lot สินค้า | `productId`, `expireDate` |
| `product_histories` | ประวัติการเคลื่อนไหว | `branchId+createdDate`, `productId` |
| `orders` | คำสั่งซื้อ | `branchId+createdDate`, `customerCode` |
| `order_items` | รายการใน order | `orderId`, `productId`, `branchId+createdDate` |
| `payments` | การชำระเงิน | `orderId`, `branchId+orderId` |
| `customers` | ลูกค้า | `code` (unique) |
| `suppliers` | ผู้จัดจำหน่าย | `clientId` |
| `receives` | การรับสินค้า | `branchId+createdDate` |
| `receive_items` | รายการรับสินค้า | `receiveId`, `lotId` |
| `stock_transfers` | โอนสต็อก | `fromBranchId+createdDate`, `toBranchId+createdDate` |
| `dispensing_logs` | บันทึกจ่ายยา | `branchId+createdDate`, `patientId+createdDate` |
| `patients` | ผู้ป่วย | `customerCode+branchId` (unique) |
| `sequences` | เลขที่เอกสาร | `type` |
| `sessions` | Redis session | - |

---

## 9. Environment Variables

```env
PORT=8586
MONGO_HOST=localhost:27017
MONGO_POS_DB_NAME=pos_db
REDIS_HOST=localhost:6379
CLIENT_ID=000
SYSTEM=POS
SECRET_KEY=your_secret_key
```

---

## 10. Non-Functional Requirements

- **Security** — JWT + Redis session, ทุก endpoint ต้องผ่าน middleware ยกเว้น public endpoints
- **Race condition prevention** — Stock transfer approve ใช้ atomic update ด้วย filter `status=PENDING`
- **Data integrity** — ลบ order/receive ต้องลบ items ที่เกี่ยวข้องด้วย
- **Audit trail** — ทุก write operation บันทึก `createdBy`, `updatedBy`, `createdDate`, `updatedDate`
- **Branch isolation** — `branchId` ฝังจาก JWT session ไม่ใช่จาก request body
- **PDF output** — ตอบกลับเป็น `application/pdf` inline (ไม่ต้อง download)
- **Excel output** — ตอบกลับเป็น `.xlsx` format

---

*สร้างจาก codebase วันที่ 19 กุมภาพันธ์ 2568*
