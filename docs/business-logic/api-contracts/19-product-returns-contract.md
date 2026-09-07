# 19. Product Returns Contract

## เป้าหมาย

กำหนด contract สำหรับการรับคืนสินค้าจาก order เดิม โดยจำกัดให้คืนได้เฉพาะจำนวนที่ผูกกับ lot จริง

## ฝั่งที่เกี่ยวข้อง

- Backend product return service
- Backend order service (อ่าน order/order item, เขียน `returnedQty`)

## Endpoints

- `POST /product-returns`
- `GET /product-returns/order/:orderId`

## Contract Expectations

### 1. Create Return

- Request ต้องมี `orderId`, `reason`, `items: [{orderItemId, quantity, refund}]` อย่างน้อยหนึ่งรายการ
- Backend ต้องตรวจทุกบรรทัดให้ผ่าน validation ทั้งหมดก่อนเริ่ม mutation ใดๆ (ไม่ apply บางส่วนแล้วค่อย reject บรรทัดถัดไป)
- เพดานคืนต่อบรรทัด = `realLotQuantity(order item) − returnedQty เดิม` ไม่ใช่ `quantity − returnedQty` เฉยๆ
- Response คืนค่า `ProductReturn` document ที่มี `returnNo` (เลขที่อ้างอิง `RT-...`), รายการที่คืนสำเร็จ, และ `totalRefund`

### 2. List Returns by Order

- คืนรายการคืนสินค้าทั้งหมดของ order นั้น เรียงจากล่าสุดไปเก่าสุด

## Error Cases

- `RT-400-001` invalid request body (รวมถึง quantity ≤ 0)
- `RT-400-002` create/query failed — ครอบคลุม: order ไม่พบ/คนละ branch, order item ไม่พบ/ไม่ใช่ของ order ที่ระบุ, ปริมาณคืนเกิน `realLotQuantity` ที่เหลือ, คืน stock กลับ lot ไม่สำเร็จ
- `RT-500-001` internal server error
