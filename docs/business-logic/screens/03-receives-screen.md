# 03. Receives Screen

## เป้าหมายของหน้าจอ

ใช้บันทึกรับสินค้าเข้าสต็อก และดูประวัติเอกสารรับสินค้าเพื่อการตรวจสอบย้อนหลัง

## ผู้ใช้เป้าหมาย

- SUPER
- ADMIN

## ข้อมูลที่ต้องโหลด

- รายการเอกสารรับสินค้า
- supplier list
- product list หรือ product search data
- receive detail ตามเอกสารที่เลือก

## องค์ประกอบหลักของ UI

- receive list
- form สำหรับเพิ่มรายการรับสินค้า
- item editor สำหรับเอกสารรับสินค้า
- summary panel ของจำนวนรายการและต้นทุนรวม

## Action หลักของผู้ใช้

- สร้างเอกสารรับสินค้าใหม่
- เพิ่ม/ลบ/แก้ item ในเอกสาร
- ยืนยันบันทึกเอกสาร
- เปิดดูประวัติการรับสินค้า

## Validation และ Feedback

- จำนวน ราคาทุน lot และวันหมดอายุ ต้องถูก validate ก่อน submit
- ถ้ามีหลายหน่วย ระบบต้องแสดงจำนวนที่แปลงเป็นหน่วยฐาน
- ถ้า save ไม่สำเร็จ ต้องระบุว่าผิดที่รายการใดหรือเงื่อนไขใด

## Empty / Loading / Error State

- หากยังไม่มีประวัติรับสินค้า ให้แสดง empty state ที่เชิญชวนให้สร้างเอกสารแรก
- loading state ของ form กับ list ควรแยกกัน

## ความสัมพันธ์กับ Backend

- Backend สร้าง receive document, stock movement, lot update และ product history
- Frontend รับผิดชอบ workflow การกรอกหลายรายการและ preview ก่อน commit
