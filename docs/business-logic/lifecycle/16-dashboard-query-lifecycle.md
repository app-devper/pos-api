# 16. Dashboard Query Lifecycle

## เป้าหมาย

อธิบาย lifecycle ของ dashboard data set ในฐานะข้อมูล derived view ที่ถูก query ใหม่จาก source-of-truth ทุกครั้ง

## Lifecycle States

### 1. `REQUESTED`

- frontend เรียก dashboard endpoint พร้อม branch context และ query params

### 2. `COMPUTED`

- backend aggregate ข้อมูลจาก orders, stocks หรือ product data ตามกฎ business ปัจจุบัน

### 3. `DISPLAYED`

- frontend นำผลลัพธ์ไป render เป็น KPI, chart, table หรือ alert widget

## State Transitions

### `REQUESTED -> COMPUTED`

- เกิดเมื่อ backend validate params และ query source data สำเร็จ

### `COMPUTED -> DISPLAYED`

- เกิดเมื่อ frontend รับ response สำเร็จและ render widget

## Business Rules

- order analytics ต้องอิงเฉพาะ order ที่ `ACTIVE`
- branch scope ต้องตรงกับ branch ปัจจุบันเสมอ
- inventory widgets ต้องไม่รวมข้าม unit แบบผิดความหมาย
- dashboard เป็น derived data จึงไม่มี persisted state ของตัวเองแบบ entity อื่น

## Failure Conditions

- invalid params เช่น threshold/days/date range → reject
- branch scope ไม่ถูกต้อง → reject
- source data ไม่พร้อมหรือ query พัง → widget นั้นต้อง fail ชัดเจน

## Expected Outcome

- dashboard เป็นภาพรวมที่สะท้อนข้อมูลจริงล่าสุดของ branch
- semantics ของ dashboard สอดคล้องกับ report/export
