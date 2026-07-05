# 10. Branches Contract

## เป้าหมาย

กำหนด contract สำหรับการจัดการข้อมูลสาขาและข้อมูลที่ frontend ใช้สร้าง branch context ในระบบ

## ฝั่งที่เกี่ยวข้อง

- Frontend หน้า branches และ flow ที่ต้องเลือกสาขา
- Backend branch services

## Contract Expectations

### 1. Branch List / Detail

- Frontend คาดหวังรายการสาขาพร้อมสถานะและข้อมูลหลัก
- Detail ควรมีข้อมูลเพียงพอสำหรับการแก้ไขหรืออ้างอิงใน settings และ workflow อื่น

### 2. Create / Update Branch

- Request ต้องมีข้อมูลสาขาพื้นฐานที่ระบบต้องใช้
- Backend ต้อง validate uniqueness และความครบถ้วนของข้อมูลสำคัญ
- Response ควรทำให้ frontend refresh list/detail ได้ทันที

### 3. Active / Inactive Handling

- Frontend คาดหวังการแยกสถานะสาขาที่ใช้งานได้กับสาขาที่ไม่ควรใช้ใน flow ใหม่
- Backend ต้องบังคับไม่ให้ branch context ที่ไม่ถูกต้องเข้าถึงข้อมูลธุรกิจที่ไม่ควรเห็น

## Error Cases

- duplicate branch code/name ตามขอบเขตที่กำหนด
- invalid branch state for requested operation
- branch not found
