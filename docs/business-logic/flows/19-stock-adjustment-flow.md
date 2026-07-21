# 19. Stock Adjustment Flow

## เป้าหมาย

ปรับยอด stock ของ lot ใดๆ พร้อมเหตุผลที่ตรวจสอบย้อนหลังได้ และ reconcile หนี้ oversell ที่ค้างอยู่โดยอัตโนมัติเมื่อปรับเพิ่ม

## Actors

- ADMIN
- SUPER
- Backend stock adjustment service

## Preconditions

- ผู้ใช้มีสิทธิ์จัดการ stock (ADMIN/SUPER)
- มี branch context ที่ถูกต้อง
- lot (`ProductStock`) ที่จะปรับมีอยู่จริงและอยู่ใน branch เดียวกับผู้ทำรายการ

## Main Flow

1. ผู้ใช้ระบุ `productId`, `stockId`, `reason`, `delta` (และ `note` ถ้ามี)
2. Backend ตรวจว่า `reason` อยู่ในชุดที่กำหนดไว้ และ `delta != 0`
3. Backend โหลด lot ปัจจุบันเพื่อเก็บ `before`
4. Backend ปรับ `quantity`: เพิ่มถ้า `delta > 0`, ลดถ้า `delta < 0` (ลดไม่ได้ถ้า stock ไม่พอ)
5. Backend สร้าง product history ของการปรับ
6. Backend บันทึก `StockAdjustment` document (`before`/`after`/`delta`/`reason`)
7. ถ้า `delta > 0` Backend ไล่ reconcile `oversoldQty` ที่ค้างของสินค้านี้แบบ FIFO โดยใช้ marker สังเคราะห์ `ADJUST:<reason>`
8. ตอบผลลัพธ์ adjustment ที่สร้างสำเร็จ

## Error Flow

- reason ไม่อยู่ในชุดที่กำหนด → reject
- delta = 0 → reject
- lot ไม่พบ หรืออยู่คนละ branch → reject
- ลดจำนวนเกินกว่าที่มีอยู่จริง → reject โดยไม่แก้ไข stock

## Expected Outcome

- stock ของ lot ที่ระบุถูกต้องตาม `before + delta = after`
- มี audit record ครบทุกครั้งที่ปรับ
- oversold liability ของสินค้าเดียวกันลดลงถ้ามีและ delta เป็นบวก
