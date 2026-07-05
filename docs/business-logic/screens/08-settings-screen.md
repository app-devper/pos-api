# 08. Settings Screen

## เป้าหมายของหน้าจอ

ใช้จัดการค่าตั้งค่าระบบและ feature toggles ที่มีผลต่อพฤติกรรมของเอกสาร หน้าจอ และ workflow ของระบบ

## ผู้ใช้เป้าหมาย

- ADMIN
- SUPER

## ข้อมูลที่ต้องโหลด

- company / branch settings
- receipt / document settings
- feature toggles
- key-value configuration ที่จำเป็นต่อระบบ

## องค์ประกอบหลักของ UI

- settings sections แยกตามหมวด
- forms สำหรับแก้ไขค่า
- save actions และ feedback area

## Action หลักของผู้ใช้

- ดูค่าปัจจุบัน
- แก้ไขข้อมูลบริษัท/สาขา
- เปิด/ปิด feature toggles
- บันทึกและตรวจสอบผลลัพธ์

## Validation และ Feedback

- ค่า required ต้องครบก่อน save
- ค่าที่มีรูปแบบเฉพาะ เช่น เบอร์โทร อีเมล ต้องตรวจเบื้องต้น
- เมื่อ save สำเร็จ ต้องสื่อว่ามีผลเมื่อใด และกระทบส่วนใดของระบบ

## Empty / Loading / Error State

- loading state ต้องรองรับกรณีโหลด settings ช้า
- หากบาง setting ไม่พร้อมใช้ ควรแสดง fallback ที่ชัดเจน

## ความสัมพันธ์กับ Backend

- Backend จัดเก็บและคืนค่า settings ตามสาขาหรือระบบ
- Frontend รับผิดชอบการจัดหมวดหมู่, ทำให้ผู้ใช้เข้าใจผลของแต่ละค่า และตัดสินใจว่าจะซ่อน/แสดง feature ใดตาม toggles ที่ได้รับ
