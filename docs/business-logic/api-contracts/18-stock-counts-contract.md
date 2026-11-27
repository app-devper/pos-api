# 18. Stock Counts Contract

## เป้าหมาย

กำหนด contract สำหรับการนับสต็อกจริงเป็นชุด แล้วให้ backend สร้าง stock adjustment ให้อัตโนมัติเฉพาะรายการที่มีผลต่าง

## ฝั่งที่เกี่ยวข้อง

- Backend stock count service
- Backend stock adjustment service (เรียกใช้ภายใน ไม่ expose แยก)

## Endpoints

- `POST /stock-counts`
- `GET /stock-counts`
- `GET /stock-counts/:id`

## Contract Expectations

### 1. Create Stock Count

- Request ต้องมี `items: [{productId, stockId, counted}]` อย่างน้อยหนึ่งรายการ (`note` เป็น optional)
- Response คืนค่า `StockCount` document ที่มี `countNo` (เลขที่อ้างอิง `SC-...`) และทุกบรรทัดที่นับ พร้อม `systemQuantity`, `countedQuantity`, `delta` ต่อบรรทัด
- บรรทัดที่ `delta != 0` ต้องมี stock adjustment ถูกสร้างขึ้นจริงเบื้องหลัง (ตรวจสอบย้อนหลังได้จาก stock adjustment list ของสินค้านั้น)
- บรรทัดที่ `delta = 0` ไม่สร้าง adjustment ใดๆ
- adjustment ของบรรทัดมี `reason` = `นับสต็อก` และ `note` = `<countNo> <note>`; product history ของบรรทัดนั้นต่อ ` — <countNo> <note>` ท้ายข้อความตามกติกา note ของ contract 17 เช่น `ปรับสต็อกสินค้า (นับสต็อก) จำนวน -2 ชิ้น — SC-0001 รอบเย็น` (ไม่มี note → ` — SC-0001`)

### 2. List / Get Stock Counts

- List คืนรายการนับทั้งหมดของ branch เรียงจากล่าสุดไปเก่าสุด
- Get by id คืนรายละเอียดทุกบรรทัดของการนับครั้งนั้น
- ทั้ง List และ Get by id: ทุกบรรทัดมี `productName`, `unit`, `lotNumber` เพิ่มจาก `productId`, `stockId` (field เพิ่มเติม ไม่มี field เดิมใดหายไป) — ระบบหาชื่อตอนอ่าน: อ่าน stock count ด้วย find เดิม (branch เดียวกัน เรียงใหม่ไปเก่า) แล้วรวม `productId` / `stockId` ของทุกบรรทัดในทุกใบ ค้นด้วย `_id` (`$in`, ใช้ index `_id`) ครั้งเดียวต่อ collection ต่อคำขอ — `products`, `product_stocks` แล้ว `product_units` ของ unit บน stock ที่พบ — ไม่ว่ามีกี่ใบ ไม่ได้เก็บไว้ในเอกสาร stock count จึงใช้กับการนับที่บันทึกก่อนการเปลี่ยนแปลงนี้ด้วย; สินค้าที่ถูก soft delete ยังคืนชื่อ; บรรทัดที่สินค้าหรือ stock ไม่อยู่แล้วยังคืนอยู่ โดยค่าที่หาไม่ได้เป็น string ว่าง `""`
- Response ของ `POST /stock-counts` คงเดิม (ไม่มีสาม field นี้)

## Error Cases

- `SC-400-001` invalid request body
- `SC-400-002` create/query failed — ครอบคลุม: stock ในบรรทัดใดบรรทัดหนึ่งไม่พบ, adjustment ของบรรทัดใดบรรทัดหนึ่งล้มเหลว
- `SC-500-001` internal server error

## Recording guarantee

- ทุกการเปลี่ยน Stock, history, เอกสาร และ reconciliation ภายในคำขอ commit ใน transaction เดียว; หากขั้นใดล้มเหลวจะ rollback ทั้งหมด
- Request/response shape และ endpoint เดิมคงเดิม; transaction retry ภายในไม่ใช่การ deduplicate HTTP request ใหม่
