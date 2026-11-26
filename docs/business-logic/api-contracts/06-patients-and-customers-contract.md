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
- `GET /customers` ค่าเริ่มต้นคืนลูกค้าทุกคน รวมคนที่ `ARCHIVED` (การลบลูกค้า = ตั้ง `status: ARCHIVED`; client แสดงพร้อมเครื่องหมาย)
- `excludeArchived=true` (opt-in) ตัดลูกค้าที่ `status` เป็น `ARCHIVED` ออก; `ACTIVE`, `INACTIVE`, `DRAFT`, ว่าง หรือไม่มี field `status` (legacy) ยังอยู่
- `excludeArchived` รับ `true` / `false` (ไม่สนตัวพิมพ์, อ่านด้วย `request.BoolQuery`); ค่าอื่น ค่าว่าง หรือส่งซ้ำ → `400 CU-400-001`
- `GET /customers/code/:customerCode` ยังหาลูกค้าที่ `ARCHIVED` เจอเสมอ (บิลเก่าอ้าง code นี้อยู่)

### 2. Patient List / Detail / Medical Data

- Frontend คาดหวังข้อมูลผู้ป่วยที่แยกส่วนข้อมูลทั่วไปกับข้อมูลสุขภาพได้
- Backend ต้อง enforce สิทธิ์เข้าถึงข้อมูลสุขภาพ
- Response ควรมีข้อมูลพอสำหรับ allergy checks, order history และ patient context ใน POS
- patient record ที่ถูก archive ไม่ควรถูกคืนใน list และ customer-code lookup ปกติ แต่ยังต้องคงอยู่เชิงประวัติ — ข้อนี้ใช้กับ **patient** record (list และ lookup ของผู้ป่วย) เท่านั้น ไม่ใช่กับ **customer**: ลูกค้าที่ `ARCHIVED` ยังอยู่ใน `GET /customers` ตามค่าเริ่มต้น (ตัดออกได้ด้วย `excludeArchived=true`) และ `GET /customers/code/:customerCode` หาเจอเสมอ ตาม §1

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
