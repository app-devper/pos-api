# 08. Promotions Contract

## เป้าหมาย

กำหนด contract สำหรับการสร้าง แก้ไข เปิด/ปิด และอ่านข้อมูลโปรโมชันที่ `pos-web` ใช้ทั้งในหน้าจัดการโปรโมชันและใน workflow การขาย

## ฝั่งที่เกี่ยวข้อง

- Frontend หน้า promotions และ POS
- Backend promotion services

## Contract Expectations

### 1. Promotion List / Detail

- Frontend คาดหวังรายการโปรโมชันที่ filter ตามสถานะหรือช่วงเวลาได้
- Detail ต้องมีข้อมูล code, name, type, value, date range, status และเงื่อนไขที่เกี่ยวข้อง

### 2. Create / Update Promotion

- Request ต้องรองรับประเภทโปรโมชัน เงื่อนไขขั้นต่ำ และขอบเขตสินค้าที่ร่วมรายการ
- Backend ต้อง validate code, date range, value, และ business constraints อื่น
- Validation error ควรแปลงกลับไปแสดงบน form ได้

### 3. Activate / Inactivate Promotion

- Frontend คาดหวังการเปลี่ยนสถานะที่สะท้อนผลกับ POS ได้ตรง
- Backend ต้องเป็น source of truth ของสถานะจริงและการหมดอายุ

### 4. POS Consumption

- POS คาดหวังข้อมูล promotion ที่ active และใช้ได้จริงกับ cart ปัจจุบัน
- Backend ต้อง reject การใช้โปรโมชันที่ inactive, expired หรือไม่ตรงเงื่อนไข

## Error Cases

- duplicate promotion code
- invalid date range
- invalid discount configuration
- apply promotion not allowed for current cart
