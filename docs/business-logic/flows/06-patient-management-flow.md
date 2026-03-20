# 06. Patient Management Flow

## เป้าหมาย

จัดการข้อมูลผู้ป่วยและข้อมูลสุขภาพเพื่อใช้ใน workflow ทางคลินิกและการจ่ายยาอย่างปลอดภัย

## Actors

- เภสัชกร
- พนักงานขายที่มีสิทธิ์
- Frontend หน้า patients และ POS
- Backend patient services

## Preconditions

- เปิดใช้งาน patient feature
- ผู้ใช้มีสิทธิ์เข้าถึงข้อมูลผู้ป่วย

## Main Flow

1. ผู้ใช้เปิดหน้าผู้ป่วย
2. Frontend โหลด patient list และข้อมูลที่จำเป็น
3. ผู้ใช้ค้นหา สร้าง หรือแก้ไขผู้ป่วย
4. ผู้ใช้บันทึก allergy, chronic disease และข้อมูลสุขภาพอื่น
5. Backend บันทึกข้อมูลและคืนผลล่าสุด
6. ผู้ใช้เลือกผู้ป่วยไปใช้ใน POS
7. Frontend และ Backend ใช้ข้อมูลผู้ป่วยในการตรวจสอบ allergy / history / dispensing context

## Error Flow

- feature ปิดอยู่ → ไม่ควรเข้าถึง flow นี้
- ไม่มีสิทธิ์ → แสดง restricted state
- ข้อมูลสุขภาพไม่ครบ → เตือนผู้ใช้และไม่หลอกว่าผ่านการตรวจครบแล้ว

## Expected Outcome

- ข้อมูลผู้ป่วยพร้อมใช้กับ POS และประวัติย้อนหลัง
- ความปลอดภัยทางยาเพิ่มขึ้นจากข้อมูลที่พร้อมใช้งาน
