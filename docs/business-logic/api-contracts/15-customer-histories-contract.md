# 15. Customer Histories Contract

## เป้าหมาย

กำหนด contract สำหรับการบันทึกและอ่านประวัติการซื้อของลูกค้า เพื่อใช้ในหน้าจัดการลูกค้าและรายงานประวัติลูกค้า

## ฝั่งที่เกี่ยวข้อง

- Frontend หน้า customers (history panel)
- Backend customer history services

## Contract Expectations

### 1. Create Customer History

- ถูกสร้างเมื่อ order ที่ผูก customer code ถูกยืนยัน
- Request ต้องมี customer code, order reference และข้อมูลสรุปที่จำเป็น

### 2. Get Customer History by Customer Code

- Frontend คาดหวังรายการประวัติการซื้อของลูกค้าเฉพาะราย
- Response ควรมีข้อมูล order date, total, items summary เพื่อแสดงใน history panel
- ใช้สำหรับแสดงบนหน้า customers และพิมพ์รายงานประวัติลูกค้า

## Endpoints

| Method | Path | คำอธิบาย |
|---|---|---|
| POST | /customer-histories | สร้างประวัติการซื้อ |
| GET | /customer-histories/:customerCode | ดูประวัติการซื้อของลูกค้า |

## Error Cases

- customer not found
- invalid customer code
- branch context ไม่ถูกต้อง
