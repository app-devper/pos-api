# 04. POS Orders Contract

## เป้าหมาย

กำหนด contract สำหรับการสร้าง order, การชำระเงิน, และข้อมูลที่เกี่ยวข้องกับการขายหน้าร้าน

## ฝั่งที่เกี่ยวข้อง

- Frontend หน้า POS
- Backend order, payment, stock deduction, dispensing และ document-related services

## Contract Expectations

### 1. Product Lookup for POS

- Frontend คาดหวังข้อมูลที่ค้นหาเร็วพอสำหรับงานหน้าร้าน
- ผลลัพธ์ควรมีข้อมูลสินค้า หน่วย ราคา และข้อมูลเพียงพอสำหรับเพิ่มเข้าตะกร้า

### 2. Submit Order

- Request ต้องรวม cart items, payment data, customer/patient references และ compliance data เมื่อจำเป็น
- Backend ต้อง validate stock, payment total, drug compliance fields และ branch authorization
- Response ควรมีข้อมูล order ที่บันทึกสำเร็จและ identifiers ที่ใช้กับเอกสารหลังการขาย

### 3. Post-Order Data

- Frontend คาดหวังข้อมูลอ้างอิงสำหรับ receipt, tax invoice, labels หรือ report-related follow-up actions
- Backend ต้องสะท้อนสถานะคำสั่งขายที่ชัดเจน

## Error Cases

- stock insufficient
- payment total invalid
- controlled drug data incomplete
- patient/customer reference invalid
- session expired during submit
