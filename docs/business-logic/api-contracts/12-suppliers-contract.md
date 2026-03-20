# 12. Suppliers Contract

## เป้าหมาย

กำหนด contract สำหรับการจัดการผู้จำหน่ายและการเลือก supplier ใน receive workflow

## ฝั่งที่เกี่ยวข้อง

- Frontend หน้า suppliers และ receives
- Backend supplier services

## Contract Expectations

### 1. Supplier List / Detail

- Frontend คาดหวังรายการ supplier ที่เลือกใช้ใน receive flow ได้สะดวก
- Detail ควรมีข้อมูลติดต่อและสถานะที่จำเป็นต่อการตัดสินใจใช้งาน

### 2. Create / Update Supplier

- Request ต้องมีข้อมูลบังคับขั้นต่ำ เช่น ชื่อและข้อมูลสำคัญที่ระบบใช้
- Backend ต้อง validate รูปแบบข้อมูลและป้องกันข้อมูลซ้ำตามกฎธุรกิจ

### 3. Supplier Usage in Receives

- Receive flow คาดหวังให้ backend ปฏิเสธ supplier ที่ invalid หรือ inactive เมื่อนโยบายไม่อนุญาต
- Response error ควร map กลับสู่ receive form ได้เข้าใจง่าย

## Error Cases

- duplicate supplier
- invalid contact data
- inactive supplier used in new receive
- supplier not found
