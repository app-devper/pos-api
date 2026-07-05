# 14. Branch Management Flow

## เป้าหมาย

จัดการสาขาให้มีข้อมูล master ที่ชัดเจน พร้อมใช้เป็น boundary หลักของ authorization และ inventory segregation

## Actors

- SUPER / ADMIN
- Frontend หน้าจัดการสาขา
- Backend branch services

## Preconditions

- ผู้ใช้มีสิทธิ์ระดับ admin/super

## Main Flow

1. ผู้ใช้เปิดหน้า branch management
2. ระบบโหลด branch list และ detail
3. ผู้ใช้สร้าง แก้ไข หรือเปลี่ยนสถานะสาขา
4. Backend generate code เมื่อสร้างสาขาใหม่
5. ข้อมูล branch ถูกใช้ต่อใน employee mapping, stock scope และ dashboard/report scope

## Error Flow

- sequence ไม่พร้อม → ห้ามสร้าง branch ด้วย code ว่าง
- branchId ไม่ถูกต้อง → reject get/update/delete
- การปิดใช้งานสาขาต้องไม่ทำให้ข้อมูลย้อนหลังหาย

## Expected Outcome

- branch master ถูกต้องและพร้อมใช้เป็น security boundary ของระบบ
