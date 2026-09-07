# 02. POS Order Flow

## เป้าหมาย

ขายสินค้าให้ลูกค้าอย่างรวดเร็ว ปลอดภัย และครบถ้วนทั้งด้านการชำระเงิน การตัดสต็อก และข้อกำหนดของยาควบคุม

## Actors

- พนักงานขาย
- เภสัชกร
- ลูกค้า / ผู้ป่วย
- Frontend หน้า POS
- Backend order / stock logic

## Preconditions

- ผู้ใช้ login สำเร็จและมี branch context
- สินค้ามี stock พร้อมขาย
- ข้อมูลผู้ป่วย/ลูกค้าโหลดได้เมื่อจำเป็น

## Main Flow

### Step 1: ค้นหาหรือสแกนสินค้า

- ผู้ใช้ค้นหาด้วยชื่อการค้า ชื่อสามัญ หรือบาร์โค้ด
- Frontend แสดงรายการที่ match อย่างรวดเร็ว
- หากสแกนเจอบาร์โค้ดของหน่วยเฉพาะ ระบบเลือกหน่วยนั้นอัตโนมัติ

### Step 2: เพิ่มเข้าตะกร้า

- ถ้าสินค้ามีหลายหน่วย Frontend แสดงตัวเลือกหน่วย
- ผู้ใช้กำหนดจำนวน
- ระบบคำนวณราคาและยอดย่อยของแต่ละบรรทัด

### Step 3: ตรวจสอบความปลอดภัย

- หากมี patient context ระบบเช็ค allergy และ drug interaction
- หากตะกร้ามีสินค้าที่ `drugRegistrations[]` มี KHY10 / KHY11 / KHY12 / KHY13 → ระบบเปิด ComplianceDialog บังคับกรอกข้อมูลเภสัชกร, เลขใบอนุญาต, ผู้ซื้อ, เลขบัตรประชาชนผู้ซื้อ และผู้สั่งจ่าย (ถ้ามี)
- **หมายเหตุ**: ดูรายละเอียด flow ข.ย. ทั้งหมดใน flow 10
- Frontend แสดง warning/block ตามเงื่อนไข

### Step 4: Review ตะกร้า

- ผู้ใช้ตรวจรายการทั้งหมด ส่วนลด โปรโมชัน และยอดสุทธิ
- ผู้ใช้สามารถเพิ่ม/ลบ/แก้จำนวนก่อนชำระเงิน

### Step 5: Payment

- Frontend เปิด payment flow
- ช่องทางชำระที่รองรับปัจจุบัน: **CASH** (เงินสด) และ **PROMPTPAY** (พร้อมเพย์)
- ผู้ใช้สามารถเพิ่มหลายแถวชำระเงินในออเดอร์เดียวกันได้ (split payment)
- ระบบคำนวณยอดรับรวมและเงินทอนแบบ real-time
- ปุ่มยืนยันชำระจะเปิดได้ต่อเมื่อยอดรับเพียงพอ

### Step 6: Submit Order

- Frontend ส่ง order payload ไปยัง Backend
- Backend ตรวจ branch context, payment totals และความสามารถในการตัด stock จาก `item.Stocks` ที่ client ส่งมา
- Compliance data ถูกบันทึกตาม payload ที่ frontend ส่งมา โดย backend ไม่ได้ derive จาก `drugRegistrations` ซ้ำในชั้นนี้

### Step 7: Commit Business Transaction

- Backend สร้าง order (รวมข้อมูล compliance: pharmacistName, licenseNo, buyerName, buyerIdCard, prescriberName)
- ตัด stock ตาม `item.Stocks` ที่ frontend เลือกและส่งมา
- อัปเดต lot balance
- สร้าง product history
- ถ้า stock movement หรือ product history ไม่ครบ ต้อง rollback เชิงธุรกิจและไม่ตอบ success
- ข้อมูลรายงาน ข.ย. 10–13 ดึงจาก order entity โดยตรง

### Step 8: Post-Order Actions

- Frontend แสดงผลสำเร็จ
- ผู้ใช้สามารถพิมพ์ receipt / tax invoice / drug label ได้
- dashboard และ report ที่เกี่ยวข้องต้องสะท้อนข้อมูลนี้ในรอบถัดไป

## Decision Points

- guest sale หรือ patient-linked sale
- มีโปรโมชันหรือไม่
- เป็น controlled drug หรือไม่
- รับชำระหลายช่องทางหรือไม่

## Error Flow

- stock ไม่พอ → reject ก่อน commit
- allergy / compliance data ไม่ครบ (ชื่อเภสัชกร, เลขใบอนุญาต, ผู้ซื้อ, เลขบัตรประชาชน) → ปัจจุบันต้อง block ที่ frontend ตาม workflow ของ POS
- payment total ไม่พอ → ไม่ให้ submit
- Backend validation fail → ไม่สร้าง order ครึ่งทาง
- stock deduction หรือ product history step ล้มเหลว → rollback order flow และตอบ error
- cancel request ที่ส่ง JSON body ผิดรูปแบบ → reject เป็น bad request ก่อนเริ่ม cancel

## Expected Outcome

- order ถูกสร้างสมบูรณ์
- stock ถูกตัดตาม stock lots ที่ frontend ส่งมา
- ข้อมูลกำกับของยาควบคุมถูกบันทึกตาม payload ที่ frontend ส่งมา
- เอกสารหลังการขายสามารถออกได้จากข้อมูลจริง
