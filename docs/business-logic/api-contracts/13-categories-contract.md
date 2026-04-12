# 13. Categories Contract

## เป้าหมาย

กำหนด contract สำหรับการจัดการหมวดหมู่สินค้าและการใช้งาน category ใน product workflows และ filters ต่าง ๆ

## ฝั่งที่เกี่ยวข้อง

- Frontend หน้า categories และ products
- Backend category services

## Contract Expectations

### 1. Category List / Detail

- Frontend คาดหวัง category list ที่ใช้ได้ทั้งในหน้าจัดการและ dropdown/filter ของสินค้า
- Detail ควรมีข้อมูลชื่อ สถานะ และ metadata ที่จำเป็น
- record ที่ถูก archive ไม่ควรถูกคืนใน list ปกติ แม้ยังต้องคงอยู่เพื่ออ้างอิงย้อนหลัง

### 2. Create / Update Category

- Request ต้องรองรับการสร้างและแก้ไขชื่อ/สถานะหมวดหมู่
- Backend ต้อง validate uniqueness และการใช้งานร่วมกับสินค้าเดิม
- การลบจากหน้าจัดการควรตีความเป็น archive semantics มากกว่า hard delete

### 3. Category Usage in Products

- Product form คาดหวังให้ category ที่ active เท่านั้นแสดงใน selector ปกติ
- Backend ต้องป้องกันการอ้างอิง category ที่ไม่ถูกต้องใน product update/create

## Error Cases

- duplicate category name
- category not found
- inactive category used in unsupported flow
- invalid reference from product form
- archived category omitted from normal list
