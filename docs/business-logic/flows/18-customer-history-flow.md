# 18. Customer History Flow

## เป้าหมาย

บันทึกและเรียกดู activity ของลูกค้าตาม branch เพื่อรองรับ CRM, review ย้อนหลัง และการอ้างอิงเชิงธุรกิจ

## Actors

- พนักงานหน้าร้าน
- ผู้จัดการสาขา
- Frontend หน้าประวัติลูกค้า
- Backend customer history services

## Preconditions

- ผู้ใช้มี branch context ถูกต้อง
- customer code ต้องอ้างอิงลูกค้าที่ระบบรู้จัก

## Main Flow

1. Frontend เปิดหน้าประวัติลูกค้า
2. ผู้ใช้เพิ่ม activity note หรือเรียกดู history ตาม customer code
3. Backend persist history พร้อม `branchId` และ `createdBy`
4. การค้นย้อนหลังดึงข้อมูลเฉพาะ branch ปัจจุบัน

## Error Flow

- bind request ไม่ผ่าน → reject
- customer code หรือ branch context ไม่ถูกต้อง → reject

## Expected Outcome

- customer history ถูกบันทึกและอ่านกลับได้อย่าง branch-scoped
- activity log พร้อมใช้ในงาน CRM และการติดตามลูกค้า
