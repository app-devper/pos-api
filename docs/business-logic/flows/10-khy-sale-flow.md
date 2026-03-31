# 10. KHY Sale Flow — การขายสินค้าที่เกี่ยวข้องกับรายงาน ข.ย.

## เป้าหมาย

อธิบาย flow การขายสินค้าที่เข้าข่ายรายงาน ข.ย.9–13 ตั้งแต่การตั้งค่าสินค้า การขายหน้าร้าน จนถึงการออกรายงาน เพื่อให้เข้าใจว่าข้อมูลไหลจากจุดไหนไปจุดไหน

## แนวคิดหลัก: `product.DrugRegistrations[]`

ระบบใช้ `product.drugRegistrations` array เป็นตัวกำหนดทั้ง **Compliance Dialog ตอนขาย** และ **การปรากฏในรายงาน ข.ย.**

### ค่าที่ใช้ได้

| DrugRegistrations | ความหมาย | Compliance Dialog |
|---|---|---|
| `KHY9` | บัญชีการซื้อยา (ผูกผ่านรับสินค้า) | ❌ ไม่ trigger |
| `KHY10` | บัญชีการขายยาควบคุมพิเศษ | ✅ ต้องกรอก |
| `KHY11` | บัญชีการขายยาอันตราย | ✅ ต้องกรอก |
| `KHY12` | บัญชีการขายยาตามใบสั่งฯ | ✅ ต้องกรอก |
| `KHY13` | รายงานการขายยาตาม อย. | ✅ ต้องกรอก |

- สินค้า 1 ตัวสามารถอยู่ในหลายรายงานพร้อมกัน
- **Trigger compliance**: ถ้าตะกร้ามีสินค้าที่ `drugRegistrations` มี KHY10 / KHY11 / KHY12 / KHY13 อย่างน้อย 1 ตัว → เปิด ComplianceDialog
- KHY9 ไม่ trigger เพราะเป็นรายงานซื้อยา ไม่ใช่ขายยา

### `product.DrugInfo.DrugType` (ข้อมูลเสริม)

- ค่า: `CONTROLLED`, `DANGEROUS`, `PSYCHO`, `NARCOTIC`
- เป็น metadata ประกอบสินค้า ใช้แสดงในรายงาน
- **ไม่ได้ใช้ trigger Compliance Dialog**

## Flow A: การตั้งค่าสินค้า (Product Setup)

### Step 1: ตั้งค่า DrugInfo

- ผู้ดูแลกำหนด `drugInfo.drugType` (CONTROLLED/DANGEROUS/PSYCHO/NARCOTIC)
- กำหนดข้อมูลยาอื่น: `genericName`, `strength`, `dosageForm`, `dosage`

### Step 2: ตั้งค่า DrugRegistrations

- ผู้ดูแลเลือกบัญชี ข.ย. ที่สินค้าต้องปรากฏ (KHY9–KHY13)
- UI แสดง checkbox สำหรับแต่ละบัญชี
- สามารถเลือกหลายบัญชีได้พร้อมกัน

### Step 3: บันทึก

- Backend เก็บ `drugRegistrations` เป็น string array ใน product document
- `drugInfo` เก็บเป็น embedded object

## Flow B: การขายหน้าร้าน (POS Sale)

### Step 1: เพิ่มสินค้าเข้าตะกร้า

- ผู้ใช้ค้นหาและเพิ่มสินค้า (รวมยาที่ผูก ข.ย.)
- ไม่มีการตรวจ DrugRegistrations ตอนเพิ่มเข้าตะกร้า
- `productRef` ถูกแนบไว้ใน cart item เพื่อใช้ตรวจสอบภายหลัง

### Step 2: กดชำระเงิน → ตรวจ Compliance

- Frontend ตรวจ `controlledDrugCompliance` feature flag
- วนตะกร้าดูว่ามีสินค้าที่ `productRef.drugRegistrations[]` มี KHY10 / KHY11 / KHY12 / KHY13 หรือไม่
- **ถ้ามี** → เปิด ComplianceDialog บังคับกรอก:
  - ชื่อเภสัชกรผู้จ่ายยา * (`pharmacistName`)
  - เลขใบอนุญาต * (`licenseNo`)
  - ชื่อผู้สั่งจ่าย (ถ้ามี) (`prescriberName`)
  - ชื่อผู้ซื้อ * (`buyerName`)
  - เลขบัตรประชาชนผู้ซื้อ * (`buyerIdCard`) — ต้องเป็นรูปแบบ X-XXXX-XXXXX-XX-X
- **ถ้าไม่มี** → ข้ามไปเปิด Payment Dialog ได้เลย

### Step 3: ยืนยันชำระเงิน

- Frontend ส่ง order payload รวม compliance fields ไป Backend
- payload fields: `pharmacistName`, `licenseNo`, `prescriberName`, `buyerName`, `buyerIdCard`
- ถ้าไม่มียาควบคุม fields เหล่านี้จะเป็น `undefined`

### Step 4: Backend สร้าง Order

- Backend สร้าง order record พร้อม compliance data
- ตัด stock ตาม FEFO
- สร้าง product history
- **compliance data ถูกเก็บใน order entity โดยตรง**

## Flow C: การออกรายงาน ข.ย. (KHY Report)

### ข.ย.9 — บัญชีการซื้อยา

- **Data Source**: `receives` (เอกสารรับสินค้า)
- **Filter**: `product.DrugRegistrations` contains "KHY9"
- **Endpoint**: `GET /reports/pharmacy/khy9/data?startDate=...&endDate=...`
- **ข้อมูลที่แสดง**: วันที่ซื้อ, เลขที่ใบรับ, ชื่อยา, ชื่อสามัญ, ความแรง, ล็อต, วันหมดอายุ, จำนวน, หน่วย, ต้นทุน, ผู้จำหน่าย

### ข.ย.10 — บัญชีการขายยาควบคุมพิเศษ

- **Data Source**: `orders` (order_items joined with products)
- **Filter**: `product.DrugRegistrations` contains "KHY10", order status = ACTIVE
- **Endpoint**: `GET /reports/pharmacy/khy10/data?startDate=...&endDate=...`
- **ข้อมูลที่แสดง**: วันที่ขาย, ชื่อยา, ชื่อสามัญ, ความแรง, รูปแบบยา, จำนวน, หน่วย, วิธีใช้, เภสัชกรผู้จ่ายยา, เลขใบอนุญาต
- **PharmacistName / LicenseNo**: ดึงจาก `order.pharmacistName`, `order.licenseNo`

### ข.ย.11 — บัญชีการขายยาอันตราย

- เหมือน ข.ย.10 แต่ filter `DrugRegistrations` contains "KHY11"
- **Endpoint**: `GET /reports/pharmacy/khy11/data?startDate=...&endDate=...`
- **หมายเหตุ**: ยาอันตราย (DANGEROUS) ไม่ trigger compliance dialog → `pharmacistName`/`licenseNo` อาจเป็นค่าว่างใน order

### ข.ย.12 — บัญชีการขายยาตามใบสั่งของผู้ประกอบวิชาชีพ

- เหมือน ข.ย.10 แต่ filter `DrugRegistrations` contains "KHY12"
- **Endpoint**: `GET /reports/pharmacy/khy12/data?startDate=...&endDate=...`

### ข.ย.13 — รายงานการขายยาตามที่เลขาธิการ อย. กำหนด

- เหมือน ข.ย.10 แต่ filter `DrugRegistrations` contains "KHY13"
- **Endpoint**: `GET /reports/pharmacy/khy13/data?startDate=...&endDate=...`

## Frontend PDF Generation

- Frontend เรียก data endpoint → ได้ `PharmacyReportResponse`
- `printPharmacyReport()` สร้าง HTML document ตาม report key:
  - KHY9: ตาราง purchase (12 columns)
  - KHY10-13: ตาราง dispensing (11 columns)
- แสดง totals, signature area, และเปิด browser print dialog

## Authorization

- ทุก endpoint ต้องผ่าน: `RequireAuthenticated` → `RequireSession` → `RequireBranch` → `RequireAuthorization(ADMIN, SUPER)`
- ข้อมูลถูกกรองตาม `branchId` จาก session

## Edge Cases และข้อควรระวัง

### 1. สินค้ามี DrugRegistrations (KHY10-13) → ได้ compliance ครบ

- สินค้าที่มี `DrugRegistrations = ["KHY10"]` / `["KHY11"]` / `["KHY12"]` / `["KHY13"]`
- จะ trigger ComplianceDialog ตอนขาย
- รายงาน ข.ย. จะมีข้อมูลเภสัชกร/ผู้ซื้อครบ

### 2. สินค้าไม่มี DrugRegistrations

- ไม่ trigger compliance dialog
- ไม่ปรากฏในรายงาน ข.ย. ใดๆ

### 3. สินค้ามีเฉพาะ KHY9

- ไม่ trigger compliance dialog (KHY9 = รายงานซื้อ ไม่ใช่ขาย)
- ปรากฏในรายงาน ข.ย.9 จากข้อมูลรับสินค้า

### 4. Order ที่ถูกยกเลิก

- รายงาน ข.ย.10-13 กรองเฉพาะ order ที่มี status = `ACTIVE`
- Order ที่ถูก void/cancel จะไม่ปรากฏในรายงาน

### 5. Compliance data ถูกแชร์ทั้ง order

- Compliance data (pharmacist, licenseNo, buyer) เก็บที่ระดับ order ไม่ใช่ order item
- ถ้ามีสินค้า KHY10 + KHY12 ใน order เดียวกัน จะใช้ข้อมูลเภสัชกรชุดเดียวกัน

## Sequence Diagram (Simplified)

```
ผู้ดูแล → [ตั้งค่า DrugType + DrugRegistrations ใน Product]
                           ↓
พนักงานขาย → [เพิ่มสินค้าเข้าตะกร้า]
                           ↓
          [ตรวจ drugRegistrations → มี KHY10-13?]
              ↓ Yes                          ↓ No
    [ComplianceDialog]              [Payment Dialog]
    [กรอก pharmacist,               [ชำระเงิน]
     licenseNo, buyer]                    ↓
              ↓                     [สร้าง Order
    [Payment Dialog]                 ไม่มี compliance]
    [ชำระเงิน]                            ↓
              ↓                     [Order ถูกบันทึก]
    [สร้าง Order
     + compliance data]
              ↓
    [Order ถูกบันทึก]
              ↓
ผู้ดูแล → [ออกรายงาน ข.ย. ตามช่วงวันที่]
              ↓
    [Backend query orders/receives]
    [Filter ตาม DrugRegistrations]
              ↓
    [Frontend render report PDF]
```
