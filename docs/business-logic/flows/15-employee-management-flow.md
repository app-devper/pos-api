# 15. Employee Management Flow

## เป้าหมาย

เชื่อม employee เข้ากับ branch และ role เพื่อให้ session ของผู้ใช้ถูกตีความเป็นสิทธิ์ใช้งานจริงในระบบ POS

## Actors

- SUPER / ADMIN
- Frontend หน้าพนักงาน
- Backend employee services

## Preconditions

- ผู้ใช้ต้องมีสิทธิ์ระดับ admin/super สำหรับ create/update/delete

## Main Flow

1. ผู้ใช้เปิดหน้า employee management
2. Frontend โหลด employee list ทั้งระบบหรือราย branch
3. ผู้ดูแลสร้างหรืออัปเดต employee profile พร้อม branch และ role
4. Backend persist employee mapping
5. Middleware authorization และ branch context ใช้ข้อมูล employee นี้ในทุก feature ต่อไป

## Error Flow

- employee mapping ไม่ถูกต้อง → session อาจไม่ตีความ branch/role ได้
- employeeId หรือ branchId ไม่ถูกต้อง → reject read/update/delete

## Expected Outcome

- ผู้ใช้แต่ละคนถูก map เข้ากับ branch และ role ที่ backend ใช้ enforce ได้จริง
