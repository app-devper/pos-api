# 03. Report Generation Flow

## เป้าหมาย

ออกรายงานและเอกสารให้ตรงกับข้อมูลจริงในระบบ โดยแยกความรับผิดชอบระหว่าง Backend และ Frontend ชัดเจน

## Actors

- ผู้ใช้ SUPER / ADMIN / USER ตามประเภทรายงาน
- Frontend หน้า reports และหน้าที่เกี่ยวข้อง
- Backend report data endpoints / CSV endpoints

## Preconditions

- ผู้ใช้มีสิทธิ์ดูรายงานนั้น
- มี input ที่จำเป็น เช่น date range, orderId, productId, customerCode

## Main Flow A: Frontend-Generated PDF

### Step 1: ผู้ใช้เลือกประเภทรายงาน

- Frontend แสดงฟอร์มกรอก date range หรือ identifier ที่จำเป็น
- ปุ่ม export จะพร้อมใช้งานเมื่อ input ครบ

### Step 2: Frontend โหลดข้อมูลดิบ

- Frontend เรียก data endpoint หรือ endpoint domain data ที่เกี่ยวข้อง
- Backend ตรวจสิทธิ์และคืนข้อมูลที่อยู่ใน branch context ที่ถูกต้อง

### Step 3: Frontend แปลงข้อมูลเป็น document model

- Frontend map ข้อมูลดิบเป็นหัวเอกสาร summary table totals และข้อความกำกับ
- ถ้าไม่มีข้อมูล ระบบแสดง empty state หรือเอกสารเปล่าแบบควบคุมได้

### Step 4: Frontend render document

- Frontend สร้าง HTML document สำหรับ preview / print
- ใช้ browser print dialog เพื่อ Save as PDF

## Main Flow B: Backend CSV Export

### Step 1: ผู้ใช้กด CSV

- Frontend เรียก CSV endpoint พร้อม params ที่จำเป็น

### Step 2: Backend สร้างไฟล์

- Backend query ข้อมูล, filter ตามสิทธิ์, format ตามโครงสร้าง CSV
- ส่งไฟล์กลับให้ browser download

## รายงาน KHY

- KHY PDF ใช้ frontend-generated document จาก `/reports/pharmacy/khy*/data`
- KHY CSV ยังคงสร้างจาก backend endpoint
- ข้อมูล KHY 9 มาจาก receives (การรับยาเข้า)
- ข้อมูล KHY 10-13 มาจาก orders (การขายยา) โดยตรง
- การคัดกรองใช้ `product.DrugRegistrations` array (KHY9, KHY10, KHY11, KHY12, KHY13)
- Order entity มี LicenseNo และ PharmacistName สำหรับรายงาน

## Error Flow

- params ไม่ครบ → Frontend block หรือ Backend reject
- ผู้ใช้ไม่มีสิทธิ์ → Backend deny
- ไม่พบข้อมูล → Frontend แสดง empty state ที่ชัดเจน
- print window ถูก block → Frontend ต้องแจ้งผู้ใช้

## Expected Outcome

- เอกสารถูกสร้างจากข้อมูลล่าสุด
- ลดภาระ backend ด้าน PDF rendering
- CSV และ data endpoint ยังตรวจสอบสิทธิ์จาก backend ได้ครบ
