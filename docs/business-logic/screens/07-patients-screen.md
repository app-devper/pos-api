# 07. Patients Screen

## เป้าหมายของหน้าจอ

ใช้จัดการข้อมูลผู้ป่วย ข้อมูลสุขภาพ และประวัติการจ่ายยา เพื่อรองรับการดูแลผู้ป่วยและความปลอดภัยทางยา

## ผู้ใช้เป้าหมาย

- USER ที่มีสิทธิ์
- ADMIN
- SUPER
- เภสัชกร

## ข้อมูลที่ต้องโหลด

- patient list
- patient detail
- allergy / chronic disease data
- patient-linked order history
- refill-related information หากมี

## องค์ประกอบหลักของ UI

- patient list / search
- detail page หรือ detail panel
- sections สำหรับข้อมูลสุขภาพ
- history section

## Action หลักของผู้ใช้

- ค้นหาและเลือกผู้ป่วย
- สร้าง/แก้ไขข้อมูลผู้ป่วย
- บันทึก allergy / chronic disease
- ดูประวัติการจ่ายยา
- เลือกผู้ป่วยไปใช้ต่อใน POS

## Validation และ Feedback

- ต้องมีการแยก field ข้อมูลทั่วไปกับข้อมูลสุขภาพอย่างชัดเจน
- ข้อมูลแพ้ยาควรระบุได้ชัดว่าผู้ป่วยแพ้อะไรและระดับใด
- ถ้าปิด patient feature ฝั่ง frontend ต้องไม่เปิด flow นี้ให้ผู้ใช้

## Empty / Loading / Error State

- ถ้าไม่มีผู้ป่วย ให้แสดง empty state พร้อม CTA
- loading ของ detail และ history ต้องไม่บังข้อมูลทั้งหมดเกินจำเป็น
- ถ้าไม่มีสิทธิ์ดูข้อมูลสุขภาพ ต้องแสดง restricted state

## ความสัมพันธ์กับ Backend

- Backend ให้ patient CRUD, medical data และประวัติที่ผูกกับผู้ป่วย
- Frontend รับผิดชอบการจัดวางข้อมูลสุขภาพให้เข้าใจง่ายและปลอดภัยต่อการใช้งาน
