# 17. Stock Adjustments Contract

## เป้าหมาย

กำหนด contract สำหรับการปรับ stock ของ lot ใดๆ แบบมีเหตุผลระบุ พร้อม audit trail ที่ตรวจสอบย้อนหลังได้

## ฝั่งที่เกี่ยวข้อง

- Backend stock adjustment service
- Backend order service (สำหรับ oversell reconciliation)

## Endpoints

- `POST /stock-adjustments`
- `GET /stock-adjustments/product/:productId`

## Contract Expectations

### 1. Create Adjustment

- Request ต้องมี `productId`, `stockId`, `reason`, `delta` (`note` เป็น optional)
- `reason` ต้องอยู่ในชุดที่กำหนด: `นับสต็อก`, `ยาเสียหาย`, `ยาหมดอายุ`, `สูญหาย`, `อื่นๆ`
- `delta` ต้องไม่เป็น 0 — บวกคือเพิ่ม stock, ลบคือลด stock
- Response คืนค่า adjustment record ที่มี `before`, `after`, `delta`, `code` (เลขที่อ้างอิง `AJ-...`)
- ถ้า `delta > 0` และมี order ที่ oversold ค้างของสินค้านี้อยู่ใน branch เดียวกัน ระบบต้อง reconcile หนี้นั้นให้อัตโนมัติในคำขอเดียวกัน (ไม่ต้องเรียกซ้ำ)

### 2. List Adjustments by Product

- คืนรายการ adjustment ทั้งหมดของสินค้านั้น เรียงจากล่าสุดไปเก่าสุด กรองตาม branch context

## Error Cases

- `AJ-400-001` invalid request body (bind ไม่ผ่าน)
- `AJ-400-002` create/query failed — ครอบคลุม: reason ไม่ถูกต้อง, delta = 0, stock ไม่พบ, stock คนละ branch, ลด stock เกินที่มีอยู่จริง
- `AJ-500-001` internal server error
