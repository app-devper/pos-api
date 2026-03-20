# 07. Customer Lifecycle

## วัตถุประสงค์

กำหนดสถานะของลูกค้าในระบบเพื่อรองรับการขาย การติดตามประวัติ และการวิเคราะห์เชิงพาณิชย์ โดยแยกบทบาทจาก patient domain ให้ชัดเจน

## สถานะหลัก

- `DRAFT` — ข้อมูลกำลังกรอกหรือแก้ไข
- `ACTIVE` — ลูกค้าพร้อมใช้งานใน order, history, และ CRM flow
- `INACTIVE` — ไม่ควรถูกใช้ในรายการใหม่ แต่ยังอ้างอิงย้อนหลังได้
- `ARCHIVED` — เก็บไว้เพื่อประวัติ ไม่ใช้กับ flow ปัจจุบัน

## Transition หลัก

### 1. `DRAFT -> ACTIVE`

#### Trigger

- ผู้ใช้บันทึกลูกค้าและข้อมูลขั้นต่ำครบ

#### Preconditions

- ชื่อหรือข้อมูลระบุตัวตนขั้นต่ำครบตามนโยบาย
- ข้อมูลติดต่ออยู่ในรูปแบบที่ยอมรับได้ถ้ามีการกรอก

#### Backend Effects

- ลูกค้าพร้อมใช้งานในหน้า POS, customer history และ segmentation ต่าง ๆ

#### Frontend Behavior

- ลูกค้าปรากฏใน search และ selector ปกติ

### 2. `ACTIVE -> INACTIVE`

#### Trigger

- ปิดใช้งานลูกค้า เช่น ไม่ต้องการใช้ในงานปัจจุบัน

#### Effects

- ไม่ควรถูกเลือกในการขายใหม่จาก quick flow
- ประวัติคำสั่งซื้อเดิมยังต้องอ้างอิงลูกค้านี้ได้

#### Frontend Behavior

- แสดงสถานะ inactive ใน list และ detail
- อาจซ่อนจาก selector ปกติ

### 3. `INACTIVE -> ACTIVE`

#### Trigger

- เปิดใช้งานใหม่

### 4. `ACTIVE/INACTIVE -> ARCHIVED`

#### Trigger

- เก็บ record เพื่ออ้างอิงระยะยาว

## ข้อควรระวัง

- ไม่ควร hard delete ลูกค้าที่ถูกใช้ใน order history แล้ว
- การรวม duplicate customers ต้องระวังไม่ให้ history สูญหายหรืออ้างอิงผิดคน
