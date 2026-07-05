# 05. POS Screen

## เป้าหมายของหน้าจอ

ทำให้การขายหน้าร้านรวดเร็ว แม่นยำ และปลอดภัย โดยรองรับทั้งงานขายทั่วไปและงานขายยาที่มีเงื่อนไขพิเศษ

## ผู้ใช้เป้าหมาย

- USER
- ADMIN
- SUPER
- เภสัชกรที่ทำงานร่วมกับการขาย

## ข้อมูลที่ต้องโหลด

- product search data
- cart state
- patient/customer lookup data
- promotion data ตามกรณีใช้งาน
- payment-related configuration หากมี

## องค์ประกอบหลักของ UI

- search area
- product result list
- cart panel
- dialogs สำหรับเลือกหน่วย, payment, patient/customer, controlled drug data
- post-order actions เช่น print receipt / label

## Action หลักของผู้ใช้

- ค้นหาหรือสแกนสินค้า
- เพิ่มสินค้าเข้าตะกร้า
- เลือกผู้ป่วย/ลูกค้า
- กรอกข้อมูลยาควบคุมเมื่อจำเป็น
- ชำระเงินและยืนยันคำสั่งขาย
- พิมพ์เอกสารหลังการขาย

## Validation และ Feedback

- quantity, stock, payment total และข้อมูลกำกับต้องถูกตรวจทั้งก่อน submit และจาก backend response
- warning ด้าน allergy หรือ interaction ต้องเด่นกว่าข้อมูลทั่วไป
- หลังบันทึกสำเร็จ ต้องสื่อให้ชัดว่ารายการนี้กลายเป็นยอดขายจริงแล้ว

## Empty / Loading / Error State

- ถ้ายังไม่มีสินค้าในตะกร้า ต้องมี empty state ที่พร้อมเริ่มขายทันที
- search loading และ submit loading ต้องแยกกัน
- ถ้า order fail ต้องรักษา state ที่ผู้ใช้สามารถแก้ไขและลองใหม่ได้

## ความสัมพันธ์กับ Backend

- Backend รับผิดชอบ order commit, stock deduction, compliance data และเอกสารข้อมูลหลังการขาย
- Frontend รับผิดชอบ interaction ความเร็วสูง, cart UX, payment UX, และ post-order print flow
