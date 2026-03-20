# 06. Customers Screen

## เป้าหมายของหน้าจอ

ใช้จัดการข้อมูลลูกค้า ค้นหาประวัติการซื้อ และรองรับการเลือก customer context ใน workflow การขาย

## ผู้ใช้เป้าหมาย

- USER
- ADMIN
- SUPER

## ข้อมูลที่ต้องโหลด

- customer list
- search/filter metadata
- customer detail summary
- purchase history ของลูกค้าที่เลือก

## องค์ประกอบหลักของ UI

- customer table หรือ list
- search input และ filters
- create/edit dialog หรือ page
- detail/history panel

## Action หลักของผู้ใช้

- ค้นหาและเปิดดูลูกค้า
- สร้าง/แก้ไขข้อมูลลูกค้า
- ดูประวัติการซื้อ
- เลือกลูกค้าไปใช้ต่อใน POS

## Validation และ Feedback

- ข้อมูลพื้นฐานที่จำเป็นต้องครบก่อน save
- ถ้าพบข้อมูลซ้ำ เช่น ชื่อ/เบอร์โทร ควรมีการเตือนก่อนสร้าง record ใหม่
- การบันทึกสำเร็จควรสะท้อนใน list และ detail ทันที

## Empty / Loading / Error State

- ถ้ายังไม่มีลูกค้า ให้แสดง empty state พร้อม CTA สำหรับสร้างรายการแรก
- loading ของ list และ history ควรแยกกัน

## ความสัมพันธ์กับ Backend

- Backend ให้ customer CRUD และ purchase history
- Frontend รับผิดชอบ search UX, detail presentation และการเลือกใช้ใน flow อื่น
