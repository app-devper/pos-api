# 15. Customer History Lifecycle

## เป้าหมาย

อธิบายสถานะของ customer history ในฐานะ CRM activity record ที่ถูกสร้างและใช้อ้างอิงย้อนหลังตามสาขา

## Lifecycle States

### 1. `ABSENT`

- ยังไม่มี history สำหรับ customer code ใน branch นั้น

### 2. `LOGGED`

- มี history record ถูกสร้างพร้อม `customerCode`, `branchId`, `createdBy`

### 3. `REVIEWED`

- history ถูกดึงไปแสดงในหน้าประวัติลูกค้าหรือใช้ในการติดตามบริการ

## State Transitions

### `ABSENT -> LOGGED`

- เกิดเมื่อผู้ใช้สร้าง customer history entry ใหม่

### `LOGGED -> REVIEWED`

- เกิดเมื่อมีการ query history ของลูกค้าใน branch นั้น

## Business Rules

- history ต้องแยกตาม `branchId`
- customer code ต้องเป็นตัวอ้างอิงหลักตอนอ่านย้อนหลัง
- history เป็น append-style log มากกว่าจะเป็น state ที่ถูกแก้ทับบ่อย

## Failure Conditions

- bind request ไม่ผ่าน → reject
- branch context หรือ customer code ไม่ถูกต้อง → reject

## Expected Outcome

- CRM activity ถูกเก็บและเรียกดูย้อนหลังได้โดยไม่ปนข้ามสาขา
