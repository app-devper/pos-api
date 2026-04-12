# 02. Receive Lifecycle

## วัตถุประสงค์

กำหนดสถานะของเอกสารรับสินค้าเพื่อให้การเพิ่ม stock และการตรวจสอบย้อนหลังมีจุดอ้างอิงที่ชัดเจน

## สถานะหลัก

- `DRAFT` — เอกสารถูกกรอกอยู่ ยังไม่เพิ่ม stock จริง
- `CONFIRMED` — เอกสารถูกบันทึกสำเร็จและเพิ่ม stock แล้ว
- `CANCELLED` — เอกสารถูกยกเลิกก่อนหรือหลังบันทึกตามนโยบายที่รองรับ

## Transition หลัก

### 1. `DRAFT -> CONFIRMED`

#### Trigger

- ผู้ใช้ยืนยันบันทึกรับสินค้า

#### Preconditions

- รายการรับเข้าครบถ้วน
- quantity และ cost price ถูกต้อง
- lot / expiry ถูกต้องเมื่อเป็นสินค้าที่ต้องติดตาม
- ผู้ใช้มีสิทธิ์และ branch ถูกต้อง

#### Backend Effects

- สร้าง receive document
- สร้าง receive items records ที่สอดคล้องกับรายการในเอกสาร
- เพิ่ม stock ตาม base quantity
- สร้างหรืออัปเดต lot
- สร้าง product history
- ทำให้ข้อมูลพร้อมใช้กับ receive summary และ KHY9

#### Frontend Behavior

- แสดงผลสำเร็จและเลขที่เอกสาร
- refresh รายการรับสินค้าและ stock ที่เกี่ยวข้อง

### 2. `DRAFT -> CANCELLED`

#### Trigger

- ผู้ใช้ยกเลิกก่อนบันทึก

#### Effects

- ไม่เพิ่ม stock
- ไม่สร้าง receive document จริง

### 3. `CONFIRMED -> CANCELLED`

#### Trigger

- ยกเลิกรายการรับสินค้าในภายหลังตามสิทธิ์และนโยบายที่ระบบรองรับ

#### Preconditions

- ต้องจัดการ reversal ของ stock และ lot ได้อย่างปลอดภัย
- ต้องไม่กระทบข้อมูล downstream โดยไร้การควบคุม

#### Backend Effects

- reverse stock / lot / history หากธุรกิจอนุญาต
- รายงานที่อิง receive ต้องตีความสถานะนี้ให้ถูก
- หากระบบยังไม่รองรับ reversal ครบถ้วน ต้องปฏิเสธการ cancel เอกสารที่ถูก import เข้า stock แล้ว แทนการลบเอกสารทิ้ง
- หากแก้ไขรายการสินค้าในช่วงที่ยังแก้ได้ ต้อง sync ทั้ง document และ receive items records ก่อนเข้าสู่ downstream flows

#### Frontend Behavior

- แสดงสถานะเอกสารว่าถูกยกเลิก
- ปิด action ที่ไม่ควรทำต่อกับเอกสารเดิม

## ข้อควรระวัง

- ห้ามเพิ่ม stock บางรายการสำเร็จแต่บางรายการล้มเหลวภายใต้เอกสารเดียวกัน
- เอกสารรับเข้าที่ `CONFIRMED` ต้องเป็นแหล่งอ้างอิงของ stock movement ที่ตรวจสอบย้อนกลับได้
- รายการสินค้าในเอกสารกับ `receive_items` collection ต้องไม่แยกคนละสถานะ เพราะรายงานและ import flow พึ่งพา collection นี้
