# 16. Supplier Management Flow

## เป้าหมาย

จัดการ supplier master และ supplier info เพื่อใช้กับ receive, product receive และรายงานจัดซื้อ

## Actors

- ผู้ดูแลข้อมูลจัดซื้อ
- Frontend หน้าซัพพลายเออร์
- Backend supplier services

## Preconditions

- ผู้ใช้ผ่าน authentication และ session แล้ว

## Main Flow

1. ผู้ใช้เปิดหน้า supplier management
2. Frontend โหลด supplier list หรือ supplier info ของ client
3. ผู้ใช้สร้าง/แก้ไข supplier หรือ supplier info
4. Backend persist supplier master
5. Receive flow และ KHY9 report ใช้ supplier data เดียวกัน

## Error Flow

- supplierId หรือ clientId ไม่ถูกต้อง → reject action
- payload ไม่ครบ → reject bind/update/create

## Expected Outcome

- supplier data พร้อมใช้ทั้งฝั่ง receive และ purchase reporting
