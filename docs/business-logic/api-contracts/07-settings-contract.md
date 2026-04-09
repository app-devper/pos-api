# 07. Settings Contract

## เป้าหมาย

กำหนด contract สำหรับการอ่านและแก้ไขค่าตั้งค่าระบบที่ `pos-web` ใช้ในการควบคุมพฤติกรรมของหน้าจอ เอกสาร และ feature toggles

## ฝั่งที่เกี่ยวข้อง

- Frontend หน้า settings และหน้าที่พึ่งพา config
- Backend settings service

## Contract Expectations

### 1. Read Settings

- Frontend คาดหวังค่า settings ที่พร้อมใช้งานได้ทันที
- Response ควรสอดคล้องกับ branch context หรือ scope ที่ระบบกำหนด

### 2. Update Settings

- Request ต้องรองรับการอัปเดตเป็นหมวดหรือเป็น key-value ตามแนวทางของระบบ
- Backend ต้อง validate ค่าที่สำคัญก่อนบันทึก
- Response ควรคืนค่าใหม่ล่าสุดหรือสถานะที่ frontend ใช้อัปเดตหน้าจอได้

### 3. Feature Toggles

- Frontend ใช้ค่าพวกนี้ในการซ่อน/แสดง flow บางส่วน
- Backend ทำหน้าที่เก็บและคืนค่า toggle configuration ของ branch
- การ enforce ว่าจะเปิด/ปิด flow ใดใน runtime ปัจจุบันเป็นความรับผิดชอบของ frontend หรือ consumer ที่อ่าน config นี้ไปใช้

## Error Cases

- unauthorized setting updates
- invalid value format
- scope mismatch ระหว่าง user กับ settings ที่กำลังแก้
