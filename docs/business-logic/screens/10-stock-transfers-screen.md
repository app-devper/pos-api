# 10. Stock Transfers Screen

## เป้าหมายของหน้าจอ

ใช้สร้าง ติดตาม และอนุมัติคำขอโอนสต็อกระหว่างสาขา โดยทำให้สถานะคำขอและผลกระทบต่อ stock ชัดเจนต่อผู้ใช้

## ผู้ใช้เป้าหมาย

- ADMIN
- SUPER

## ข้อมูลที่ต้องโหลด

- stock transfer list
- transfer detail
- branches ที่เกี่ยวข้อง
- product และ stock availability สำหรับการโอน

## องค์ประกอบหลักของ UI

- transfer list/table
- create transfer form
- detail panel
- action buttons สำหรับ approve / reject / cancel

## Action หลักของผู้ใช้

- สร้างคำขอโอน
- ดูรายละเอียดคำขอ
- อนุมัติ ปฏิเสธ หรือยกเลิกคำขอ
- ตรวจสอบสถานะล่าสุดของคำขอ

## Validation และ Feedback

- ห้ามเลือกต้นทางและปลายทางเป็นสาขาเดียวกัน
- quantity ต้องมากกว่า 0
- หน้าเว็บควรบอกเหตุผลเมื่อ approve ไม่ได้ เช่น stock ไม่พอ

## Empty / Loading / Error State

- ถ้ายังไม่มีรายการโอน ให้แสดง empty state พร้อม CTA
- loading ของ list และ detail/action ควรแยกกัน

## ความสัมพันธ์กับ Backend

- Backend รับผิดชอบ validate stock, เปลี่ยนสถานะ, และย้าย stock จริงเมื่อ approve
- Frontend รับผิดชอบ workflow การยืนยัน action และแสดงสถานะล่าสุดให้ถูกต้อง
