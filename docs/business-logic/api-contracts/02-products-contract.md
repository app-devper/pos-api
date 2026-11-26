# 02. Products Contract

## เป้าหมาย

กำหนด contract สำหรับการอ่านและจัดการข้อมูลสินค้า สต็อก หน่วยนับ และข้อมูลที่หน้า product ต้องใช้

## ฝั่งที่เกี่ยวข้อง

- Frontend หน้า products และหน้าที่ใช้ product selectors
- Backend product, stock, history, และ related services

## Contract Expectations

### 1. Product List

- Frontend คาดหวังรายการสินค้าแบบ paginated หรือ filterable
- แต่ละรายการควรมีข้อมูลเพียงพอต่อการแสดง list เช่น ชื่อ ราคา สถานะ และข้อมูล stock ระดับสรุปถ้ามี

### 2. Product Detail

- เมื่อผู้ใช้เปิดดูรายละเอียด Frontend คาดหวังข้อมูลที่ลึกขึ้น เช่น หน่วยนับ ราคา stock summary และ metadata สำคัญ

### 3. Product Create / Update

- Request ต้องมีข้อมูลบังคับตาม business rules
- Backend ต้อง validate ชื่อ หน่วย ราคา และความสอดคล้องของ multi-unit configuration
- หากข้อมูลไม่ครบหรือขัดกัน ต้องตอบ error ที่ map กลับสู่ form ได้

### 4. Stock / Unit / Lot Delete Semantics

- การลบ `product lot` ต้องถูกปฏิเสธถ้า lot นั้นยังมี `quantity > 0`
- การลบ `product stock` ต้องถูกปฏิเสธถ้า stock record นั้นยังมี `quantity > 0`
- การลบ `product unit` ต้องถูกปฏิเสธถ้ายังมี stock history หรือ stock records ที่อ้าง `unitId` นั้นอยู่
- Frontend ควรมอง error กลุ่มนี้เป็น business rule violation ไม่ใช่ transient system error

### 4.1 Product Delete (`DELETE /products/:productId`, branch admin)

- เป็น soft delete: สินค้าได้ `deletedDate` (และ `updatedBy` = ผู้ลบ) แล้วหายจาก product list, product detail และรายงานสต็อก
- สินค้าที่ยังมีสต็อกลบได้ (frontend ถามยืนยันก่อน): ใน transaction เดียวกัน ทุก `product_stocks` ของสินค้านั้นที่ `quantity != 0` ในทุกสาขาถูกปรับเป็น 0 ผ่าน Stock ledger แบบเดียวกับการปรับสต็อก
  - แต่ละ stock ได้ `stock_adjustments` หนึ่งใบ: `reason` = `อื่นๆ`, `note` = `ลบสินค้า`, `delta` = `-quantity`, `after` = 0
  - แต่ละ stock ได้ product history หนึ่งแถว: `ลบสินค้า: ปรับสต็อกเป็น 0 จำนวน <delta> <unit>` (stock ที่ขายเกินจนติดลบได้ delta เป็นบวก)
  - ถ้าขั้นใดล้มเหลว ไม่มีอะไรเปลี่ยน (สต็อก ใบปรับ ประวัติ และสินค้า)
- สินค้าที่อยู่ในใบโอนสต็อกที่ยังรออนุมัติ (`PENDING`) ลบไม่ได้: ของที่จองไว้ออกจาก stock ต้นทางไปแล้ว ถ้าลบ การ reject จะคืนของเข้า stock ของสินค้าที่ถูกลบ และการ approve จะเปิด stock ใหม่ให้สินค้าที่ถูกลบที่สาขาปลายทาง — ต้อง approve หรือ reject ใบโอนก่อนจึงลบได้; ตรวจใน transaction เดียวกันก่อน mark ลบ และถูกปฏิเสธแล้วไม่มีอะไรเปลี่ยน
- สต็อกถูกปรับเป็น 0 ในทุกสาขา ไม่ใช่เฉพาะสาขาของผู้ลบ
- Order line ที่ยังค้างส่งสินค้านี้ (`oversoldQty > 0`) ยังค้างอยู่เหมือนเดิม: การปรับ stock ที่ติดลบขึ้นเป็น 0 ไม่ถือเป็นการส่งของค้าง และไม่เขียนประวัติ "ส่งของค้างลูกค้า"
- Response: `200` พร้อม product object ที่ถูกลบ (มี `deletedDate`); `updatedBy` / `updatedDate` ของการลบถูกเขียนลงเอกสารสินค้า แต่ไม่อยู่ใน JSON response (`entities.Product` ซ่อนทั้งสอง field ด้วย `json:"-"`)
- Errors: product id ไม่ถูกต้อง → `400 PD-400-002`; สินค้าอยู่ในใบโอนที่รออนุมัติ → `400 PD-400-002` ข้อความ `สินค้านี้อยู่ในใบโอนที่รออนุมัติ ลบไม่ได้`; ไม่พบสินค้าหรือถูกลบไปแล้ว → `404 SY-404-002`; สต็อกเปลี่ยนระหว่างบันทึก → `409 SY-409-001` (ส่งซ้ำได้)

### 5. Product History / Print-related Data

- Frontend อาจต้องใช้ข้อมูลสินค้าเพื่อสร้าง price tags, price list หรือประวัติสินค้า
- Backend ต้องคืนข้อมูลดิบที่เพียงพอให้ frontend render เอกสารหรือหน้าแสดงผลต่อได้

## Error Cases

- product not found
- duplicate barcode
- invalid unit mapping
- cannot remove lot with remaining quantity
- cannot remove stock with remaining quantity
- cannot remove unit with stock history
- inactive product ถูกเรียกใช้ใน flow ที่ไม่อนุญาต
