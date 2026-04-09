# 13. Settings Lifecycle

## เป้าหมาย

อธิบายสถานะเชิงธุรกิจของ branch-level settings ตั้งแต่ยังไม่มีข้อมูล ไปจนถึงถูกสร้างและอัปเดตใช้งานจริง

## Lifecycle States

### 1. `MISSING`

- branch ยังไม่มี settings record
- ระบบอาจใช้ default behavior ภายในบางหน้าจอหรือเอกสาร

### 2. `INITIALIZED`

- มี settings record สำหรับ branch แล้ว
- fields อาจยังไม่ครบทุกส่วน แต่ record พร้อมถูกอ้างอิงโดย report / promptpay / receipt / company info

### 3. `UPDATED`

- settings ถูกแก้ไขล่าสุดโดยผู้มีสิทธิ์
- ค่าที่แก้ใหม่มีผลกับการ render เอกสารและ behavior ที่อ้าง setting ทันที

## State Transitions

### `MISSING -> INITIALIZED`

- เกิดเมื่อเรียก upsert ครั้งแรกของ branch นั้น
- ระบบสร้าง settings record ใหม่

### `INITIALIZED -> UPDATED`

- เกิดเมื่อผู้ใช้แก้ค่าผ่าน settings screen
- ระบบ persist ทับค่าเดิมของ branch เดียวกัน

## Business Rules

- settings ต้องถูก scope ตาม `branchId`
- feature ที่อ้าง settings ต้องทนต่อกรณีข้อมูลบาง field ยังไม่ถูกตั้งค่า
- การอัปเดต settings ไม่ควรกระทบข้อมูลย้อนหลังในเอกสารที่ generate ไปแล้ว เว้นแต่เป็นเอกสารที่ render realtime ใหม่ทุกครั้ง

## Failure Conditions

- branch context ไม่ถูกต้อง → ต้อง reject request
- payload ไม่สมบูรณ์หรือ parse ไม่ได้ → ต้อง reject ก่อน update

## Expected Outcome

- แต่ละ branch มี configuration ของตัวเอง
- report/document paths ใช้ค่าของ branch ปัจจุบันได้สม่ำเสมอ
