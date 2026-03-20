# 01. Auth and Context Contract

## เป้าหมาย

กำหนด contract สำหรับการพิสูจน์ตัวตน บริบทของระบบ และข้อมูลที่ `pos-web` ต้องมีเพื่อเรียก `pos-api` ได้อย่างถูกต้อง

## ผู้เกี่ยวข้อง

- `pos-web`
- `UM API`
- `pos-api`

## Contract Expectations

### 1. Login Context

- Frontend login ผ่าน UM API ก่อน
- หลัง login สำเร็จ Frontend ต้องได้ข้อมูลที่ใช้เรียก POS API ต่อได้
- Frontend ต้องมีข้อมูล host/base URL ของ POS API ก่อนยิงคำขอธุรกิจ

### 2. Token Usage

- ทุก request ที่ต้องใช้สิทธิ์ต้องแนบ token ที่ถูกต้อง
- Backend ต้อง validate token และสิทธิ์ก่อนเข้าถึงข้อมูล
- ถ้า token ไม่ถูกต้องหรือหมดอายุ ต้องตอบกลับในรูปแบบที่ Frontend แปลเป็น session-expired flow ได้

### 3. Branch Context

- Request ที่ผูกกับธุรกิจระดับสาขาต้องถูกตีความภายใต้ branch context
- Backend ต้องไม่คืนข้อมูลข้ามสาขาโดยไม่มีสิทธิ์
- Frontend ต้องถือว่า branch context เป็นส่วนหนึ่งของ state หลักของ session

## Response Expectations

- Frontend คาดหวังข้อมูลที่เพียงพอต่อการรู้ว่า session ใช้งานได้หรือไม่
- ถ้า authorization ไม่ผ่าน Backend ต้องตอบในรูปแบบที่แยกจาก validation error ทั่วไปได้

## Error Cases

- token หมดอายุ
- ไม่มีสิทธิ์เข้าถึง resource
- ไม่มี branch context ที่ใช้ได้
- host หรือ config ที่จำเป็นยังไม่พร้อมใช้งาน
