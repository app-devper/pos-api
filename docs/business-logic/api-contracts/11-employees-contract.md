# 11. Employees Contract

## เป้าหมาย

กำหนด contract สำหรับการจัดการข้อมูลพนักงานและการใช้งานข้อมูลบุคลากรในหน้า web ต่าง ๆ

## ฝั่งที่เกี่ยวข้อง

- Frontend หน้า employees
- Backend employee services

## Contract Expectations

### 1. Employee List / Detail

- Frontend คาดหวัง list ที่ค้นหาและ filter ได้
- Detail ควรมีข้อมูลส่วนบุคคลที่ระบบอนุญาตให้แสดง รวมถึงความสัมพันธ์กับสาขาและสถานะ

### 2. Create / Update Employee

- Request ต้องมีข้อมูลบังคับและความสัมพันธ์กับ branch ที่ถูกต้อง
- Backend ต้อง validate branch reference และ business constraints ที่เกี่ยวข้อง

### 3. State Handling

- Frontend คาดหวังสถานะ active/inactive ที่สะท้อนตรงกับสิทธิ์ใช้งานจริง
- Backend ต้องป้องกันการใช้งาน employee ที่ไม่พร้อมใช้งานใน flow ที่จำเป็น

## Error Cases

- invalid branch reference
- duplicate identifier ตามที่ธุรกิจกำหนด
- employee not found
- invalid state transition or unsupported action
