# 11. Supplier Lifecycle

## วัตถุประสงค์

กำหนดวงจรชีวิตของผู้จำหน่ายเพื่อให้การรับสินค้า การตรวจสอบต้นทางของสินค้า และการอ้างอิงย้อนหลังมีความถูกต้อง

## สถานะหลัก

- `DRAFT` — ข้อมูล supplier กำลังถูกสร้างหรือแก้ไข
- `ACTIVE` — พร้อมใช้งานใน receive flow และเอกสารธุรกิจที่เกี่ยวข้อง
- `INACTIVE` — ไม่ควรถูกใช้ในรายการใหม่ แต่ยังต้องอ้างอิงย้อนหลังได้
- `ARCHIVED` — เก็บไว้เพื่อประวัติระยะยาว

## Transition หลัก

### 1. `DRAFT -> ACTIVE`

#### Trigger

- ผู้ใช้บันทึก supplier พร้อมข้อมูลที่จำเป็น

#### Preconditions

- ชื่อ supplier และข้อมูลติดต่อหลักอยู่ในรูปแบบที่ใช้งานได้

#### Backend Effects

- supplier พร้อมให้เลือกใน receive document และ workflow ที่เกี่ยวข้อง

#### Frontend Behavior

- supplier ปรากฏใน selector และ list ปกติ

### 2. `ACTIVE -> INACTIVE`

#### Trigger

- ปิดใช้งาน supplier

#### Effects

- ไม่ควรถูกเลือกใน receive ใหม่
- เอกสารรับสินค้าเดิมยังต้องอ้างอิง supplier นี้ได้

#### Frontend Behavior

- แสดงสถานะ inactive
- ซ่อนจาก quick selection ปกติหรือแสดงแบบ disabled

### 3. `INACTIVE -> ACTIVE`

#### Trigger

- เปิดใช้งาน supplier อีกครั้ง

### 4. `ACTIVE/INACTIVE -> ARCHIVED`

#### Trigger

- เก็บไว้เพื่ออ้างอิงประวัติเท่านั้น

## ข้อควรระวัง

- ห้าม hard delete supplier ที่เคยถูกใช้ใน receive documents แล้ว
- การ merge supplier ซ้ำต้องระวังไม่ให้เอกสารย้อนหลังอ้างอิงผิดต้นทาง
