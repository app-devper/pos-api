# 02. POS Order Flow

## เป้าหมาย

ขายสินค้าให้ลูกค้าอย่างรวดเร็ว ปลอดภัย และครบถ้วนทั้งด้านการชำระเงิน การตัดสต็อก และข้อกำหนดของยาควบคุม

## Actors

- พนักงานขาย
- เภสัชกร
- ลูกค้า / ผู้ป่วย
- Frontend หน้า POS
- Backend order / stock / dispensing logic

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
- หากเป็นยาควบคุม ระบบบังคับกรอกข้อมูลเภสัชกร ผู้ซื้อ และข้อมูลกำกับอื่น
- Frontend แสดง warning/block ตามเงื่อนไข

### Step 4: Review ตะกร้า

- ผู้ใช้ตรวจรายการทั้งหมด ส่วนลด โปรโมชัน และยอดสุทธิ
- ผู้ใช้สามารถเพิ่ม/ลบ/แก้จำนวนก่อนชำระเงิน

### Step 5: Payment

- Frontend เปิด payment flow
- ผู้ใช้เลือกช่องทางชำระหนึ่งหรือหลายช่องทาง
- ระบบคำนวณยอดรับรวมและเงินทอนแบบ real-time
- ปุ่มยืนยันชำระจะเปิดได้ต่อเมื่อยอดรับเพียงพอ

### Step 6: Submit Order

- Frontend ส่ง order payload ไปยัง Backend
- Backend ตรวจ stock, compliance data, payment totals, และ branch context ซ้ำอีกครั้ง

### Step 7: Commit Business Transaction

- Backend สร้าง order
- ตัด stock ตาม FEFO
- อัปเดต lot balance
- สร้าง product history
- สร้าง dispensing log สำหรับรายการที่เข้าข่ายรายงานยา

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
- allergy / compliance data ไม่ครบ → block ยืนยันคำสั่งขาย
- payment total ไม่พอ → ไม่ให้ submit
- Backend validation fail → ไม่สร้าง order ครึ่งทาง

## Expected Outcome

- order ถูกสร้างสมบูรณ์
- stock ถูกตัดถูก lot
- ข้อมูลกำกับของยาควบคุมถูกบันทึกครบ
- เอกสารหลังการขายสามารถออกได้จากข้อมูลจริง
