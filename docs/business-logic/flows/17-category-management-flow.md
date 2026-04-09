# 17. Category Management Flow

## เป้าหมาย

จัดการหมวดหมู่สินค้าให้พร้อมใช้ใน product catalog, filtering และ default grouping

## Actors

- ผู้ดูแลสินค้า
- Frontend หน้าหมวดหมู่สินค้า
- Backend category services

## Preconditions

- ผู้ใช้ผ่าน authentication และ session แล้ว

## Main Flow

1. ผู้ใช้เปิดหน้า category management
2. ระบบโหลด category list และค่า default category
3. ผู้ใช้สร้าง แก้ไข ตั้ง default หรือลบ category
4. Backend persist category state
5. Product screens ใช้ category เหล่านี้ในการจัดกลุ่มและกรองข้อมูล

## Error Flow

- categoryId ไม่ถูกต้อง → reject get/update/delete/default
- payload ไม่ครบ → reject create/update

## Expected Outcome

- category master พร้อมใช้ใน product listing และ data entry flows
