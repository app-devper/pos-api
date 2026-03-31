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

### 3. Supplier Info (ข้อมูลร้านในฐานะผู้จำหน่าย)

- `PUT /suppliers/info` — อัปเดตข้อมูลร้านค้าตนเองในฐานะ supplier (เช่น ใช้ในเอกสารรับสินค้า)
- `GET /suppliers/info` — อ่านข้อมูลร้านค้าตนเองในฐานะ supplier
- ข้อมูลนี้แยกจาก supplier ทั่วไปที่เป็นคู่ค้าภายนอก

### 4. Supplier Usage in Receives

- Receive flow คาดหวังให้ backend ปฏิเสธ supplier ที่ invalid หรือ inactive เมื่อนโยบายไม่อนุญาต
- Response error ควร map กลับสู่ receive form ได้เข้าใจง่าย

## Endpoints

| Method | Path | คำอธิบาย |
|---|---|---|
| POST | /suppliers | สร้าง supplier |
| GET | /suppliers | ดูรายการ suppliers |
| GET | /suppliers/:supplierId | ดูรายละเอียด supplier |
| PUT | /suppliers/:supplierId | แก้ไข supplier |
| DELETE | /suppliers/:supplierId | ลบ supplier |
| PUT | /suppliers/info | อัปเดตข้อมูลร้านค้า (self) |
| GET | /suppliers/info | อ่านข้อมูลร้านค้า (self) |

## Error Cases

- duplicate supplier
- invalid contact data
- inactive supplier used in new receive
- supplier not found
