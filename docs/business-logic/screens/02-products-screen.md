# 02. Products Screen

## เป้าหมายของหน้าจอ

ใช้จัดการข้อมูลสินค้า ราคา หน่วยนับ และตรวจสอบความเคลื่อนไหวของสินค้าในระบบ

## ผู้ใช้เป้าหมาย

- SUPER
- ADMIN
- USER ที่มีสิทธิ์ดูหรือแก้ไขสินค้า

## ข้อมูลที่ต้องโหลด

- รายการสินค้าแบบ paginated
- หมวดหมู่สินค้า
- ข้อมูลหน่วยนับ
- stock summary ต่อสินค้า
- ข้อมูลที่ใช้พิมพ์ price tags / price list

## องค์ประกอบหลักของ UI

- ตารางรายการสินค้า
- search / filter / pagination
- ปุ่ม create / edit / view detail
- actions สำหรับพิมพ์ price tag, price list, และดู history

## Action หลักของผู้ใช้

- ค้นหาสินค้า
- สร้าง/แก้ไขสินค้า
- เปิดดูรายละเอียดสินค้า
- พิมพ์ป้ายราคาและรายการราคา
- ไปยังรายงานหรือ history ที่เกี่ยวข้อง

## Validation และ Feedback

- field สำคัญ เช่น ชื่อสินค้า ราคา หน่วยหลัก ต้องถูกตรวจสอบก่อน submit
- ถ้าสินค้ามีหลายหน่วย Frontend ต้องช่วยอธิบาย relation ของหน่วยให้เข้าใจง่าย
- การ save สำเร็จควรสะท้อนผลที่ list ทันที

## Empty / Loading / Error State

- ถ้าไม่มีสินค้าตามคำค้น ให้แสดง empty state พร้อมคำแนะนำในการค้นใหม่
- ถ้าโหลด price/stock ไม่สำเร็จ ควรแยก feedback จากการโหลด list หลัก

## ความสัมพันธ์กับ Backend

- Backend ให้ product CRUD, stock summary, history data และข้อมูลดิบสำหรับรายงาน/พิมพ์เอกสาร
- Frontend รับผิดชอบ table interaction, dialogs, และ frontend-generated printing สำหรับบางเอกสาร
