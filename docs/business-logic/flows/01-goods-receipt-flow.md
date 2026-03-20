# 01. Goods Receipt Flow

## เป้าหมาย

รับสินค้าเข้าระบบอย่างถูกต้อง พร้อมอัปเดต stock, lot, receive document และ product history อย่างสอดคล้องกัน

## Actors

- ผู้ใช้ SUPER / ADMIN
- Frontend หน้า receives / product receive flow
- Backend receive service
- Product / Lot / Stock / History repositories

## Preconditions

- ผู้ใช้มีสิทธิ์เข้าถึงหน้ารับสินค้า
- มี branch context ถูกต้อง
- สินค้าที่จะรับเข้ามีอยู่ในระบบ หรือถูกสร้างก่อนเริ่ม flow

## Main Flow

### Step 1: เปิดหน้ารับสินค้า

- Frontend โหลดข้อมูล supplier, product และข้อมูลที่จำเป็นต่อการรับเข้า
- ผู้ใช้เริ่มสร้างเอกสารรับสินค้าใหม่

### Step 2: เพิ่มรายการสินค้า

- ผู้ใช้เลือกสินค้า
- กรอกจำนวน ราคาทุน lot number วันหมดอายุ และหน่วยที่รับเข้า
- Frontend ตรวจ validation เบื้องต้น เช่น จำนวนมากกว่า 0 และข้อมูลบังคับครบ
- หากสินค้ามีหลายหน่วย ระบบคำนวณ base quantity ให้ผู้ใช้เห็นก่อนบันทึก

### Step 3: Review รายการทั้งหมด

- ผู้ใช้ตรวจรายการทั้งหมดในเอกสารรับสินค้า
- Frontend แสดงยอดรวมต้นทุนและจำนวนรายการ
- ผู้ใช้สามารถลบหรือแก้ไขรายการก่อนยืนยัน

### Step 4: Submit

- Frontend ส่งข้อมูล receive document ทั้งชุดไปยัง Backend
- Backend ตรวจสิทธิ์, branch context, โครงสร้างข้อมูล, และ business constraints

### Step 5: Persist Business Transaction

- Backend สร้าง receive document
- Backend สร้างหรืออัปเดต lot ของสินค้า
- Backend เพิ่ม stock ตาม base quantity ที่คำนวณแล้ว
- Backend สร้าง product history สำหรับแต่ละรายการ
- Backend คำนวณ total cost ระดับเอกสาร

### Step 6: Response and Refresh

- Frontend แสดงผลสำเร็จ
- หน้า list / detail refresh ด้วยข้อมูลล่าสุด
- เอกสารรับสินค้าสามารถถูกใช้ในรายงานสรุปรับสินค้าและ ข.ย.9 ต่อได้

## Decision Points

- ถ้าสินค้ามีหลายหน่วย ต้องแปลงหน่วยก่อน commit
- ถ้า lot เดิมมีอยู่แล้ว ต้องตัดสินใจตามนโยบายระบบว่าจะรวม balance หรือสร้าง record เพิ่ม
- ถ้ารายการใดรายการหนึ่ง invalid ต้อง reject ทั้งเอกสาร

## Error Flow

- ถ้าข้อมูลไม่ครบ Frontend ต้อง block ก่อน submit เท่าที่ตรวจได้
- ถ้า Backend พบข้อมูลไม่สมบูรณ์ ต้อง reject และไม่สร้างข้อมูลครึ่งทาง
- ถ้า save สำเร็จบางส่วนไม่ได้ ระบบต้อง rollback เชิงธุรกิจให้ข้อมูลไม่ค้างกลาง

## Expected Outcome

- receive document ถูกสร้างครบ
- stock และ lot balance ถูกอัปเดตถูกต้อง
- มี product history สำหรับการตรวจสอบย้อนหลัง
