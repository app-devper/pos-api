# 03. POS and Clinical Support

## วัตถุประสงค์ทางธุรกิจ

ทำให้การขายสินค้าและยาที่หน้าร้านรวดเร็ว ปลอดภัย และสอดคล้องกับข้อกำหนดทางกฎหมายและข้อมูลทางคลินิกของผู้ป่วย

## ผู้มีส่วนเกี่ยวข้อง

- พนักงานขาย
- เภสัชกร
- ลูกค้า
- ผู้ป่วย

## Workflow References

- `flows/02-pos-order-flow.md` — sequence flow ของการขายหน้าร้าน
- `flows/10-khy-sale-flow.md` — compliance flow ของ ข.ย.10–13
- `lifecycle/01-order-lifecycle.md` — order lifecycle
- `api-contracts/04-pos-orders-contract.md` — order / POS contracts

## End-to-End Workflow Summary

### 1. Build Cart

- พนักงานค้นหาสินค้าหรือสแกนบาร์โค้ด
- ระบบเลือกสินค้าและหน่วยขายที่ถูกต้อง
- quantity และ stock availability ต้องถูก validate ตั้งแต่ก่อนชำระเงิน

### 2. Attach Customer or Patient Context

- ถ้าเป็นลูกค้าทั่วไป ระบบอาจผูก customer code เพื่อใช้กับประวัติการซื้อ
- ถ้าเป็นผู้ป่วย ระบบผูก patient context เพื่อใช้กับ allergy check และข้อมูลกำกับทางคลินิก

### 3. Run Clinical Safety Checks

- เมื่อมี patient context ระบบตรวจ allergy และ clinical constraints ก่อนปิดการขาย
- warning ที่รุนแรงต้องถูกแจ้งก่อน commit order
- เป้าหมายคือ block หรือเตือนให้เร็วที่สุด ไม่ใช่หลังตัด stock ไปแล้ว

### 4. Collect Controlled-Drug Metadata

- ถ้าสินค้าอยู่ในกลุ่มที่ trigger ข.ย.10–13 ระบบต้องเก็บ pharmacist, license, buyer และข้อมูลกำกับอื่นให้ครบ
- หากข้อมูลบังคับไม่ครบ ต้องไม่ให้ยืนยัน order

### 5. Validate Payment and Commit Order

- ระบบรับ payment ได้หลายช่องทางใน order เดียว
- ยอดรวม payment ต้องครอบคลุมยอดสุทธิ
- เมื่อ validation ผ่าน Backend จึง commit order, order items, payments และ stock movement

### 6. Deduct Stock and Generate Audit Trail

- stock ถูกตัดตาม unit / quantity ของ transaction จริง
- ระบบสร้าง product history และรักษา consistency ระหว่าง order กับ inventory
- ถ้าไม่สามารถสร้าง product history ได้ ต้องถือว่า order-side stock movement ยังไม่สมบูรณ์และไม่ควรถูกมองว่า success
- ถ้าตัด stock ไม่ครบหรือ step สำคัญล้มเหลว ต้องไม่ปล่อย order ค้างครึ่งทาง

### 7. Create Downstream Clinical Outputs

- ประวัติการขายและ patient context ถูกใช้เป็นข้อมูลอ้างอิงต่อในงานติดตามผู้ป่วย
- ข้อมูลสำหรับ ข.ย.10–13 ถูกอ่านจาก order/product data ตามกฎ compliance

## Business Flow หลัก

1. ค้นหาสินค้าหรือสแกนบาร์โค้ด
2. เลือกหน่วยขาย
3. เพิ่มสินค้าเข้าตะกร้า
4. ระบบตรวจสอบ interaction / allergy / controlled drug requirement
5. เลือกผู้ป่วยหรือข้อมูลลูกค้า (ถ้ามี)
6. รับชำระเงิน
7. ยืนยันออเดอร์
8. ตัดสต็อกและบันทึกข้อมูลรายงานที่เกี่ยวข้อง

## Business Rules

### 1. Product Search and Selection

- ค้นหาสินค้าได้จากชื่อการค้า ชื่อสามัญ และบาร์โค้ด
- ถ้าสินค้ามีหลายหน่วย ระบบต้องให้เลือกหน่วยก่อนเพิ่มเข้าตะกร้า
- สินค้าที่ไม่มี stock เพียงพอไม่ควรถูกขายเกินยอดคงเหลือ

### 2. Cart Management

- ระบบรองรับหลาย cart พร้อมกันเพื่อรองรับลูกค้าหลายราย
- สินค้าเดียวกันต่างหน่วยถือเป็นคนละรายการในตะกร้าได้
- การเปลี่ยนจำนวนต้องตรวจสอบ stock ใหม่ทุกครั้ง

### 3. Clinical Safety

- ถ้ามี patient context ระบบต้องตรวจสอบ allergy และ interaction
- allergy ที่ตรงเงื่อนไขรุนแรงต้อง block หรือเตือนระดับสูง
- interaction ควรแจ้งเตือนก่อนชำระเงิน ไม่ใช่หลังปิดการขาย
- ปัจจุบันเป็น workflow ที่ frontend ต้องเรียกและ enforce ก่อน submit order

### 4. Controlled Drug Compliance

- ยากลุ่มที่กำหนดต้องบันทึกข้อมูลกำกับก่อนปิดการขาย
- **Trigger**: ตรวจจาก `product.drugRegistrations[]` มี KHY10 / KHY11 / KHY12 / KHY13 (KHY9 ไม่ trigger เพราะเป็นรายงานซื้อ)
- ใช้ `drugRegistrations` เดียวกันกับการกรองรายงาน ข.ย. (ดูรายละเอียดใน flow 10)
- ข้อมูลที่ต้องกรอก: ชื่อเภสัชกร (`pharmacistName`), เลขใบอนุญาต (`licenseNo`), ชื่อผู้ซื้อ (`buyerName`), เลขบัตรประชาชนผู้ซื้อ (`buyerIdCard`)
- ข้อมูลเสริม: ชื่อผู้สั่งจ่าย (`prescriberName`)
- เลขบัตรประชาชนต้องผ่านการตรวจรูปแบบ (X-XXXX-XXXXX-XX-X)
- หากข้อมูลบังคับไม่ครบ ต้องไม่อนุญาตให้ยืนยันออเดอร์
- ข้อมูลนี้ถูกบันทึกใน order entity โดยตรง เพื่อรองรับรายงาน ข.ย. อัตโนมัติ

### 5. Payment Logic

- รองรับหลายช่องทางการชำระในออเดอร์เดียว
- ช่องทางที่รองรับปัจจุบัน: **CASH** (เงินสด) และ **PROMPTPAY** (พร้อมเพย์)
- ผู้ใช้สามารถเพิ่มหลายแถวชำระเงินในออเดอร์เดียวกันได้ (เช่น บางส่วนเงินสด บางส่วนพร้อมเพย์)
- ยอดรวมการรับชำระต้องไม่น้อยกว่ายอดขายสุทธิ
- หากรับเกินต้องคำนวณเงินทอนให้ถูกต้อง
- การบันทึกการขายสำเร็จได้ต่อเมื่อ payment ผ่าน validation แล้ว

### 6. Stock Deduction

- การยืนยันขายใช้ stock lots ที่ frontend เลือกและส่งมาใน `item.Stocks`
- Backend ตัด stock, lot balance, และ product history จาก payload นั้นใน transaction เชิงธุรกิจเดียวกัน
- ถ้าตัด stock ไม่ครบ ต้องยกเลิกการบันทึกออเดอร์ทั้งชุด
- ถ้า unit lookup หรือ product history step ล้มเหลว ต้อง fail flow นั้นเช่นกัน เพื่อคง audit trail ให้ครบ

### 7. KHY Data Source

- เมื่อขายสินค้าที่อยู่ในกลุ่มรายงานยา ข้อมูลต้องถูกบันทึกใน order entity และข้อมูลกำกับที่เกี่ยวข้อง
- รายงาน ข.ย. 10–13 ดึงข้อมูลจาก orders (order_items joined with products) โดยตรง
- การคัดกรองใช้ `product.DrugRegistrations` array (ค่า: KHY10, KHY11, KHY12, KHY13)
- ข้อมูลเภสัชกรและเลขใบอนุญาตดึงจาก order entity

## Validation Rules

- quantity ต้องมากกว่า 0
- unit ต้อง map กับ product จริง
- buyer / pharmacist / licenseNo / prescriber data ต้องครบเมื่อสินค้าอยู่ในกลุ่มควบคุมตาม rule ฝั่ง frontend/POS workflow
- payment total ต้องครอบคลุมยอดสุทธิ

## Edge Cases

- สแกนบาร์โค้ดที่ไม่รู้จัก ต้องไม่สร้างสินค้าใหม่อัตโนมัติ
- stock เปลี่ยนระหว่างที่กำลังชำระเงิน ต้อง recheck ก่อน commit
- cart ที่ค้างไว้หลายชุดต้องไม่ปนข้อมูลกัน

## Expected Outcomes

- ออเดอร์ถูกต้อง
- stock ถูกตัดตรง
- ข้อมูลเพื่อรายงาน ข.ย. ถูกสร้างครบ
- ความเสี่ยงทางคลินิกลดลงจากการเตือนก่อนขาย
