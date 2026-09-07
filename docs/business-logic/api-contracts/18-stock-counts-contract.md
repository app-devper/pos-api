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

### 2. List / Get Stock Counts

- List คืนรายการนับทั้งหมดของ branch เรียงจากล่าสุดไปเก่าสุด
- Get by id คืนรายละเอียดทุกบรรทัดของการนับครั้งนั้น

## Error Cases

- `SC-400-001` invalid request body
- `SC-400-002` create/query failed — ครอบคลุม: stock ในบรรทัดใดบรรทัดหนึ่งไม่พบ, adjustment ของบรรทัดใดบรรทัดหนึ่งล้มเหลว
- `SC-500-001` internal server error
