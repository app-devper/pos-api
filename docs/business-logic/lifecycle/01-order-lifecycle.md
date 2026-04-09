# 01. Order Lifecycle

## วัตถุประสงค์

อธิบายสถานะของคำสั่งขายตั้งแต่เริ่มสร้างจนถึงการปิดรายการ เพื่อให้การคำนวณยอด การตัดสต็อก การออกเอกสาร และรายงานอ้างอิงสถานะเดียวกัน

## สถานะหลัก

- `DRAFT` — ตะกร้าหรือข้อมูลที่ผู้ใช้กำลังแก้ไข ยังไม่เป็นยอดขายจริง
- `CONFIRMED` — คำสั่งขายถูกบันทึกสำเร็จและถือเป็นยอดขายจริง
- `CANCELLED` หรือ `VOIDED` — รายการถูกยกเลิกและต้องไม่ถูกรวมในรายงานยอดขายปกติ

## Transition หลัก

### 1. `DRAFT -> CONFIRMED`

#### Trigger

- ผู้ใช้ยืนยันการชำระเงินและ submit order

#### Preconditions

- รายการสินค้าในตะกร้าถูกต้อง
- stock เพียงพอ
- payment total ครอบคลุมยอดสุทธิ
- ข้อมูลกำกับยาควบคุมครบถ้วนเมื่อจำเป็นตาม workflow ฝั่ง frontend
- branch context และ authorization ถูกต้อง

#### Backend Effects

- สร้าง order record (รวม compliance data: `pharmacistName`, `licenseNo`, `prescriberName`, `buyerName`, `buyerIdCard`)
- ตัด stock ตาม `item.Stocks` ที่ request ส่งมา
- อัปเดต lot balance
- สร้าง product history
- ข้อมูลรายงาน ข.ย. 10–13 ดึงจาก order entity โดยตรง
- ทำให้ข้อมูลพร้อมสำหรับ receipt, tax invoice, sales report และ KHY reports

#### Frontend Behavior

- แสดงสถานะสำเร็จพร้อมเลขที่เอกสาร
- reset หรือ close cart ตาม workflow
- เปิดทางเลือกพิมพ์ receipt / tax invoice / labels

### 2. `DRAFT -> CANCELLED`

#### Trigger

- ผู้ใช้ล้างตะกร้า หรือยกเลิกก่อน submit

#### Effects

- ไม่สร้าง order จริง
- ไม่ตัด stock
- ไม่สร้างข้อมูลรายงาน

#### Frontend Behavior

- ล้าง state ของตะกร้าอย่างชัดเจน
- ไม่ทิ้งยอดค้างที่ทำให้ผู้ใช้เข้าใจผิด

### 3. `CONFIRMED -> VOIDED` หรือ `CANCELLED`

#### Trigger

- มีการยกเลิกรายการขายหลังบันทึก ตามนโยบายธุรกิจที่อนุญาต

#### Preconditions

- ผู้ใช้มีสิทธิ์สูงพอ
- มีเหตุผลรองรับการยกเลิก
- ระบบรองรับการคืน stock หรือ reversal อย่างถูกต้อง

#### Backend Effects

- รายการต้องไม่ถูกนับในยอดขายปกติ
- หากมีการคืน stock ต้องอัปเดต stock, lot, และ history อย่างสอดคล้อง
- รายงานที่อิงยอดขายต้องตีความสถานะนี้อย่างชัดเจน

#### Frontend Behavior

- แสดง badge สถานะว่าเป็นรายการถูกยกเลิก
- ปิด action ที่ไม่ควรทำซ้ำ เช่น void ซ้ำ

## ข้อควรระวัง

- ห้ามมีสถานะกึ่งสำเร็จที่ order ถูกสร้างแต่ stock ไม่ถูกตัด
- เอกสารหลังการขายต้องอ้างอิงเฉพาะ order ที่ `CONFIRMED` หรือสถานะที่ธุรกิจกำหนด
- Dashboard และ report ต้องไม่รวม order ที่ถูกยกเลิกโดยไม่ตั้งใจ
