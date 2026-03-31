# 05. Reports Contract

## เป้าหมาย

กำหนด contract สำหรับรายงานที่ `pos-web` ใช้ทั้งในรูปแบบ data-driven preview, frontend-generated PDF และ backend CSV export

## ฝั่งที่เกี่ยวข้อง

- Frontend หน้า reports และหน้าที่มีปุ่มเอกสาร
- Backend report data endpoints และ CSV endpoints

## Contract Expectations

### 1. Data Endpoints

- Frontend คาดหวังข้อมูลดิบที่พร้อมใช้ render preview และสร้างเอกสาร
- Request มักมี date range หรือ identifiers เฉพาะรายงาน
- Backend ต้องกรองข้อมูลตามสิทธิ์และ branch context

### 2. Frontend-generated PDF Flow

- Frontend ใช้ data endpoint ชุดเดียวกับ preview เพื่อหลีกเลี่ยงข้อมูลไม่ตรงกัน
- Response ต้องมีหัวเอกสาร รายการ และ summary เท่าที่ frontend ต้องใช้ render document

### 3. CSV Export

- Frontend เรียก endpoint export โดยส่ง params แบบเดียวกับ flow รายงานนั้น
- Backend ต้องคืนไฟล์ที่มีโครงสร้างสอดคล้องกับความหมายทางธุรกิจของรายงาน

### 4. KHY Reports

- KHY 9-13 ใช้ data endpoints สำหรับ frontend PDF / preview
- CSV ยังคงมาจาก backend flow
- Backend ต้องแยกแหล่งข้อมูลให้ถูก:
  - KHY9: receives (เอกสารรับสินค้า) กรองด้วย `product.DrugRegistrations` contains "KHY9"
  - KHY10-13: orders (order_items joined with products) กรองด้วย `product.DrugRegistrations` contains "KHY10"-"KHY13"
- ข้อมูลเภสัชกร (`pharmacistName`) และเลขใบอนุญาต (`licenseNo`) ดึงจาก order entity
- Response ต้องมี: date, productName, genericName, quantity, unit, strength, dosageForm, dosage, pharmacistName, licenseNo

## Error Cases

- missing filters
- unauthorized report access
- no data in selected range
- incompatible parameters for specific report types
