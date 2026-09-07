# 21. Product Return Flow

## เป้าหมาย

รับคืนสินค้าจาก order เดิม โดยคืนได้เฉพาะจำนวนที่ผูกกับ lot จริงเท่านั้น และคืน stock กลับไปยัง lot ที่ถูกต้อง

## Actors

- ADMIN
- SUPER
- Backend product return service

## Preconditions

- ผู้ใช้มีสิทธิ์จัดการ return
- order ที่อ้างอิงอยู่ใน branch เดียวกับผู้ทำรายการ
- order item ที่จะคืนยังมีจำนวนที่ผูกกับ lot จริงเหลืออยู่ (ยังไม่ถูกคืนหมด)

## Main Flow

1. ผู้ใช้ระบุ `orderId`, เหตุผล, และรายการที่จะคืน `{orderItemId, quantity, refund}`
2. Backend โหลด order และตรวจ branch ให้ตรงกัน
3. สำหรับแต่ละบรรทัด: Backend โหลด order item และคำนวณ `realLotQuantity` (ผลรวม `stocks[]` ที่ไม่ใช่ synthetic marker)
4. Backend ตรวจว่า `returnedQty เดิม + quantity ที่จะคืน` ไม่เกิน `realLotQuantity` — ถ้าเกิน reject พร้อมข้อความระบุจำนวนคืนได้สูงสุด
5. Backend จัดสรรจำนวนที่จะคืนกลับไปยัง lot ที่ถูกต้อง โดยไล่ตามลำดับ `stocks[]` ข้ามส่วนที่ถูกคืนไปแล้วและข้าม synthetic marker เสมอ
6. Backend คืน `quantity` เข้า `ProductStock` ของแต่ละ lot ที่จัดสรรได้ และเพิ่ม `returnedQty` บน order item
7. Backend สร้าง product history ของการคืน
8. Backend บันทึก `ProductReturn` document พร้อมยอดคืนเงินรวม (`totalRefund`)

## Error Flow

- order ไม่พบ หรือคนละ branch → reject ทั้งคำขอ
- order item ไม่พบ หรือไม่ได้เป็นของ order ที่ระบุ → reject ทั้งคำขอ
- ปริมาณที่จะคืนเกิน `realLotQuantity` ที่เหลือ → reject บรรทัดนั้นก่อนเริ่ม mutation ใดๆ (ตรวจครบทุกบรรทัดก่อน apply จริง)

## Expected Outcome

- stock ของ lot เดิมกลับมาถูกต้องตามจำนวนที่คืนได้จริง
- `returnedQty` สะสมถูกต้องแม้คืนหลายครั้งบางส่วน
- ส่วนที่เป็น oversold หรือ synthetic adjustment ไม่ถูกนำมาคำนวณเป็นยอดที่คืนได้
