# 04. POS Orders Contract

## เป้าหมาย

กำหนด contract สำหรับการสร้าง order, การชำระเงิน, และข้อมูลที่เกี่ยวข้องกับการขายหน้าร้าน

## ฝั่งที่เกี่ยวข้อง

- Frontend หน้า POS
- Backend order, payment, stock deduction และ document-related services

## Contract Expectations

### 1. Product Lookup for POS

- Frontend คาดหวังข้อมูลที่ค้นหาเร็วพอสำหรับงานหน้าร้าน
- ผลลัพธ์ควรมีข้อมูลสินค้า หน่วย ราคา และข้อมูลเพียงพอสำหรับเพิ่มเข้าตะกร้า

### 2. Submit Order

- Request ต้องรวม cart items, payment data, customer/patient references และ compliance data เมื่อจำเป็น
- Compliance fields ในกรณียาควบคุม: `pharmacistName`, `licenseNo`, `prescriberName`, `buyerName`, `buyerIdCard`
- Backend ต้อง validate branch authorization, payment total และความสามารถในการตัด stock จาก `item.Stocks`
- ปัจจุบัน compliance fields ถูกเก็บตาม payload ที่ client ส่งมา และ frontend เป็นตัว enforce rule ว่าต้องกรอกเมื่อใด
- Response ควรมีข้อมูล order ที่บันทึกสำเร็จและ identifiers ที่ใช้กับเอกสารหลังการขาย
- Order entity เก็บ compliance data โดยตรง เพื่อรองรับรายงาน ข.ย. 10–13 ที่ดึงจาก orders

### 3. Post-Order Data

- Frontend คาดหวังข้อมูลอ้างอิงสำหรับ receipt, tax invoice, labels หรือ report-related follow-up actions
- Backend ต้องสะท้อนสถานะคำสั่งขายที่ชัดเจน

### 4. Order Item Management

- Frontend สามารถดึง order items แยกจาก order หลักได้
- รองรับการค้นหา items ตาม product เพื่อใช้ในรายงานหรือ history
- รองรับการลบ item ออกจาก order (ก่อน finalize)

### 5. Customer Code Update

- Frontend สามารถอัปเดต customer code ของ order ที่สร้างแล้วได้ผ่าน PATCH
- ใช้ในกรณีที่ผูกลูกค้าภายหลังการสร้าง order

## Endpoints

| Method | Path | คำอธิบาย |
|---|---|---|
| POST | /orders | สร้าง order ใหม่ |
| GET | /orders | ดูรายการ orders |
| GET | /orders/:orderId | ดูรายละเอียด order |
| DELETE | /orders/:orderId | ลบ/void order |
| PATCH | /orders/:orderId/customer-code | อัปเดต customer code |
| GET | /orders/items | ดูรายการ order items (range) |
| GET | /orders/items/:itemId | ดูรายละเอียด order item |
| DELETE | /orders/items/:itemId | ลบ order item |
| GET | /orders/items/products/:productId | ค้นหา items ตาม product |
| GET | /orders/item-details/products/:productId | ค้นหา item details ตาม product |
| GET | /orders/customer/:customerCode | ดู orders ของลูกค้า |

## Error Cases

- stock insufficient
- compliance data missing for controlled drugs
- invalid payment total
- invalid product or unit reference
- branch context ไม่ถูกต้อง
- order not found
- invalid order state for requested action
- session expired during submit
