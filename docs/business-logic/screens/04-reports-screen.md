# 04. Reports Screen

## เป้าหมายของหน้าจอ

ใช้เลือก ดู preview และ export รายงานเชิงธุรกิจและรายงานตามกฎหมายจากข้อมูลจริงในระบบ

## ผู้ใช้เป้าหมาย

- SUPER
- ADMIN
- USER สำหรับรายงานที่ไม่ถูกจำกัดเฉพาะผู้ดูแล

## ข้อมูลที่ต้องโหลด

- company / setting data ที่ใช้แสดงบนเอกสาร
- report data ตามประเภทและเงื่อนไขที่ผู้ใช้เลือก
- metadata ที่ใช้ build document เช่น date range, totals, branch context

## องค์ประกอบหลักของ UI

- รายการประเภท report
- filters เช่น date range, product, customer, order reference
- action buttons เช่น preview, CSV, PDF
- status/feedback area

## Action หลักของผู้ใช้

- เลือกประเภทรายงาน
- กำหนด filter หรือ date range
- เรียก preview
- export CSV หรือ PDF

## Validation และ Feedback

- ปุ่ม action ต้องพร้อมใช้เมื่อ input ที่จำเป็นครบแล้วเท่านั้น
- ถ้ารายงานประเภทนั้นใช้ frontend-generated PDF ต้องสื่อสาร flow ให้ชัด
- หากไม่มีข้อมูล ต้องแสดงผลแบบ empty state ไม่ใช่ popup เงียบ ๆ

## Empty / Loading / Error State

- loading state ต้องบอกว่ากำลังโหลดข้อมูลหรือกำลังสร้างเอกสาร
- error state ควรอธิบายว่าเกิดจากสิทธิ์, input, หรือ server response

## ความสัมพันธ์กับ Backend

- Backend รับผิดชอบ data endpoints และ CSV endpoints
- Frontend รับผิดชอบ preview, print, และการสร้างเอกสาร PDF สำหรับ report ที่ย้ายมาฝั่ง client
