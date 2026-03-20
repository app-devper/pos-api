# 09. Promotions Screen

## เป้าหมายของหน้าจอ

ใช้สร้าง แก้ไข ติดตาม และควบคุมสถานะโปรโมชัน เพื่อให้ผู้ใช้ธุรกิจบริหารเงื่อนไขส่วนลดได้อย่างชัดเจน

## ผู้ใช้เป้าหมาย

- ADMIN
- SUPER

## ข้อมูลที่ต้องโหลด

- promotion list
- promotion detail
- product data สำหรับเลือกสินค้าเข้าร่วมโปรโมชัน
- status และช่วงเวลาของโปรโมชัน

## องค์ประกอบหลักของ UI

- promotions table/list
- create/edit form
- status badges
- filters ตามสถานะและช่วงเวลา

## Action หลักของผู้ใช้

- สร้างโปรโมชัน
- แก้ไขเงื่อนไขโปรโมชัน
- เปิด/ปิดสถานะ
- ตรวจสอบว่าโปรโมชันใดกำลัง active หรือ expired

## Validation และ Feedback

- code ต้องไม่ว่างและไม่ซ้ำ
- start/end date ต้องสมเหตุสมผล
- value, minPurchase, maxDiscount ต้องตรวจตามประเภทโปรโมชัน
- การเปลี่ยนสถานะควรมี feedback ชัดเจน

## Empty / Loading / Error State

- ถ้ายังไม่มีโปรโมชัน ให้แสดง empty state พร้อม CTA
- หน้า list และ form ควรมี loading แยกกัน

## ความสัมพันธ์กับ Backend

- Backend ให้ promotion CRUD และ validation เชิงธุรกิจ
- Frontend รับผิดชอบ form UX, status presentation และการช่วยผู้ใช้ตั้งค่าไม่ให้ขัดกัน
