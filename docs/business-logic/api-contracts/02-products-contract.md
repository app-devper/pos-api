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
