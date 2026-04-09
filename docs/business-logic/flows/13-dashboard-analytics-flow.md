# 13. Dashboard and Analytics Flow

## เป้าหมาย

แสดงตัวชี้วัดรายสาขาจากข้อมูลจริงที่ผ่านกฎ active-status และ branch scope เดียวกับรายงานทางธุรกิจ

## Actors

- ผู้บริหารสาขา
- พนักงานที่ได้รับสิทธิ์ดู dashboard
- Frontend dashboard
- Backend analytics queries

## Preconditions

- ผู้ใช้มี branch context ถูกต้อง
- วันที่และ query params ที่ส่งมาต้อง parse ได้

## Main Flow

1. Frontend เปิด dashboard พร้อม branch ปัจจุบัน
2. Frontend เรียก summary, daily chart, monthly chart, stock report, low stock, dead stock, expiring และ refill reminders
3. Backend query เฉพาะข้อมูลของ branch ปัจจุบัน
4. Query ฝั่ง order analytics ใช้เฉพาะ order ที่ `ACTIVE`
5. Query ฝั่ง inventory แยกผลตามหน่วยที่แท้จริงเมื่อข้อมูล stock อยู่หลาย unit
6. Frontend render KPI และ charts จากชุดข้อมูลเดียวกันกับกฎใน reports

## Error Flow

- invalid query params เช่น days / threshold / date range → reject
- branch scope ไม่ถูกต้อง → reject
- upstream data inconsistency ต้องไม่ทำให้ dashboard ปนข้อมูลข้ามสาขา

## Expected Outcome

- dashboard ตรงกับ export/report semantics
- KPI ไม่รวม canceled/inactive orders
- inventory widgets ไม่ปนข้อมูลข้าม branch หรือข้าม unit
