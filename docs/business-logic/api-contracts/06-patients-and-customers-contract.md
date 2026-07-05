# 06. Patients and Customers Contract

## เป้าหมาย

กำหนด contract สำหรับข้อมูลลูกค้า ผู้ป่วย และประวัติที่ frontend ใช้ในหน้าจัดการข้อมูลและหน้า POS

## ฝั่งที่เกี่ยวข้อง

- Frontend หน้า customers, patients, POS
- Backend customer, patient และ history-related services

## Contract Expectations

### 1. Customer List / Detail / History

- Frontend คาดหวัง list ที่ค้นหาได้และ detail ที่ดูประวัติการซื้อได้
- Request และ response ต้องรองรับการใช้งานทั้งบนหน้าจัดการและการเลือกไปใช้ใน POS

### 2. Patient List / Detail / Medical Data

- Frontend คาดหวังข้อมูลผู้ป่วยที่แยกส่วนข้อมูลทั่วไปกับข้อมูลสุขภาพได้
- Backend ต้อง enforce สิทธิ์เข้าถึงข้อมูลสุขภาพ
- Response ควรมีข้อมูลพอสำหรับ allergy checks, order history และ patient context ใน POS
- patient record ที่ถูก archive ไม่ควรถูกคืนใน list และ customer-code lookup ปกติ แต่ยังต้องคงอยู่เชิงประวัติ

### 3. Create / Update Flows

- Request ต้อง map กับ form state ได้ตรง
- Validation error ควรระบุ field หรือเหตุผลที่นำกลับไปแสดงใน form ได้
- การลบ patient จากหน้าจัดการควรเก็บเป็น archive semantics แทน hard delete เพื่อไม่ทำลาย clinical history

## Error Cases

- duplicate customer/patient
- insufficient permission for medical data
- patient flow not available in current frontend feature configuration
- record not found
- archived patient omitted from normal lookup
