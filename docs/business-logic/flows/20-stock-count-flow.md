# 20. Stock Count Flow

## เป้าหมาย

นับสต็อกจริงเทียบกับระบบเป็นชุด แล้วสร้าง stock adjustment ให้อัตโนมัติเฉพาะรายการที่มีผลต่าง โดยไม่ต้องให้ผู้ใช้สร้าง adjustment เองทีละรายการ

## Actors

- ADMIN
- SUPER
- Backend stock count service (เรียกใช้ stock adjustment service ภายใน)

## Preconditions

- ผู้ใช้มีสิทธิ์จัดการ stock
- มี branch context ที่ถูกต้อง

## Main Flow

1. ผู้ใช้ส่งชุดรายการนับ `{productId, stockId, counted}` มาพร้อมกันในคำขอเดียว
2. Backend สร้างเลขที่นับ (`countNo`) จาก sequence
3. สำหรับแต่ละบรรทัด: Backend โหลด `systemQuantity` ปัจจุบันของ lot แล้วคำนวณ `delta = counted - systemQuantity`
4. ถ้า `delta != 0` Backend เรียก stock adjustment logic เดียวกับ flow 19 (เหตุผล = `นับสต็อก`) ซึ่งรวม oversell reconciliation ถ้ามี
5. ถ้า `delta = 0` ไม่มีการสร้าง adjustment สำหรับบรรทัดนั้น
6. Backend บันทึก `StockCount` document ที่รวมทุกบรรทัดที่นับ (ทั้งที่มีและไม่มีผลต่าง)
7. ตอบผลลัพธ์ stock count ที่สร้างสำเร็จ

## Error Flow

- lot ในบรรทัดใดบรรทัดหนึ่งไม่พบ → reject การนับทั้งชุด
- adjustment ของบรรทัดใดบรรทัดหนึ่งล้มเหลว (เช่น validation ของ stock adjustment ไม่ผ่าน) → reject คำขอ; รายการก่อนหน้าที่ apply ไปแล้วจะไม่ rollback อัตโนมัติ (best-effort เช่นเดียวกับ receive/stock transfer)

## Expected Outcome

- ทุกบรรทัดที่มีผลต่างมี stock adjustment record อ้างอิงเหตุผล `นับสต็อก`
- เอกสารการนับสต็อกเก็บครบทุกบรรทัดไว้เป็นหลักฐานว่านับครบ ไม่ใช่แค่บรรทัดที่ผิดปกติ
