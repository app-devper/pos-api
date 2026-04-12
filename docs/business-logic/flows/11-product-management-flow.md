# 11. Product Management Flow

## เป้าหมาย

จัดการ product master, unit, price, stock, lot และ product history ให้สอดคล้องกันในระดับสาขา และพร้อมใช้กับ receive, order, dashboard และ reports

## Actors

- SUPER / ADMIN
- Frontend หน้าสินค้าและ inventory
- Backend product services
- Product / Unit / Price / Stock / Lot / History repositories

## Preconditions

- ผู้ใช้ผ่าน authentication และ session แล้ว
- ผู้ใช้มี branch context ถูกต้องสำหรับ flow ที่กระทบ stock
- ผู้ใช้มีสิทธิ์ตาม action เช่น create/update/delete สินค้า

## Main Flows

### Flow A: Create Product Master

1. ผู้ใช้เปิดหน้าสร้างสินค้า
2. Frontend รับข้อมูลหลัก เช่น serial number, name, default unit, cost price, selling price และข้อมูลยา
3. Backend สร้าง product master พร้อม default unit, default price และ product history
4. Frontend refresh product list/detail

### Flow B: Add or Update Product Unit / Price

1. ผู้ใช้เปิดหน้า detail สินค้า
2. ผู้ใช้เพิ่มหน่วยใหม่หรือแก้ราคาของหน่วยเดิม
3. Backend ตรวจ unit mapping, price payload และสิทธิ์ของผู้ใช้
4. Backend persist unit/price พร้อม product history ใน transaction เดียว
5. หน่วยและราคาที่แก้ไขถูกใช้ต่อใน receive, order และ report

### Flow C: Manual Stock Adjustment

1. ผู้ใช้เปิด stock tab ของสินค้าในสาขาปัจจุบัน
2. ผู้ใช้เพิ่ม/แก้ quantity, lot, expire date หรือ sequence
3. Backend persist stock เฉพาะ branch ปัจจุบัน
4. Backend คำนวณ stock balance ของ branch เดียวกันและสร้าง product history
5. ถ้าโหลด unit ไม่ได้หรือสร้าง history ไม่สำเร็จ Backend ต้อง reject action
6. Dashboard / stock report สะท้อนค่าล่าสุดของสาขานั้น

### Flow D: Product Lot Management

1. ผู้ใช้ดู list lot ของสินค้า
2. ผู้ใช้ create/update/delete lot ตาม branch context
3. Backend persist branch-aware lot data
4. ถ้า lot ยังมี quantity คงเหลือ Backend ต้อง reject การ delete
5. Expire notify และ inventory screens ใช้ข้อมูล lot/stocks ของ branch เดียวกัน

### Flow E: Remove Product Unit or Stock Safely

1. ผู้ใช้สั่งลบ unit หรือ stock จากหน้าจัดการสินค้า
2. Backend โหลดข้อมูลปัจจุบันภายใต้ branch context ที่เกี่ยวข้อง
3. ถ้า stock ยังมี quantity คงเหลือ Backend ต้อง reject การลบ
4. ถ้า unit ยังมี stock records อ้างอิงอยู่ Backend ต้อง reject การลบ
5. เมื่อผ่าน validation แล้วจึงลบ metadata ที่ปลอดภัยและบันทึก history ตาม flow ปกติ

## Decision Points

- ถ้าสินค้ามีหลายหน่วย ต้องระบุ conversion logic ให้ครบก่อนใช้ใน receive/order
- ถ้าแก้ unit หรือ price ต้องไม่ทำให้รายการขาย/รับเข้าย้อนหลังสูญเสียความหมาย
- ถ้าเป็น stock adjustment ต้องสร้าง history เพื่อ audit ได้เสมอ
- ถ้าสร้าง history ไม่ได้ ต้อง fail action นั้น ไม่ใช่ log แล้วปล่อยผ่าน
- lot เก่าที่ไม่มี branchId ต้องถูกตีความตาม fallback policy ของระบบ
- ถ้าลบ lot/unit/stock แล้วทำให้ inventory reality หรือ audit trail หายความหมาย ต้อง reject action

## Error Flow

- serial number ซ้ำ → reject การสร้างสินค้า
- unit mapping ไม่ครบ → reject create/update unit
- stock operation ทำให้ quantity ติดลบ → reject action
- unit lookup หรือ product history creation สำหรับ stock movement ล้มเหลว → reject action
- delete lot/stock/unit ที่ยังมีผลกับ inventory จริง → reject action
- branch context ไม่ถูกต้อง → reject flow ที่กระทบ stock/lot/history

## Expected Outcome

- product master และ inventory structures ของสินค้าอยู่ใน state ที่พร้อมใช้งาน
- stock, lot และ history ไม่ปนข้ามสาขา
- การตรวจสอบย้อนหลังทำได้จาก product history และ stock/lot records
