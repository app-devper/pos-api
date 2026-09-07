# 09. Stock Adjustments, Stock Counts, and Product Returns

## วัตถุประสงค์ทางธุรกิจ

รองรับการแก้ไขยอด stock ที่ไม่ได้มาจาก order หรือ receive โดยตรง (นับสต็อกจริง, ของเสีย/หมดอายุ/สูญหาย, ลูกค้าคืนสินค้า) พร้อมทั้งรองรับการขาย "oversell" แบบมีการติดตามหนี้ (oversold liability) และ reconcile หนี้นั้นอัตโนมัติเมื่อมี stock ใหม่เข้ามา ไม่ว่าจะจาก receive import หรือจาก stock adjustment/count เอง

## Workflow References

- `flows/19-stock-adjustment-flow.md` — ปรับสต็อกแบบมีเหตุผลระบุ (manual adjustment)
- `flows/20-stock-count-flow.md` — นับสต็อกจริงแล้วสร้าง adjustment อัตโนมัติต่อรายการที่ผลต่าง
- `flows/21-product-return-flow.md` — รับคืนสินค้าที่ผูกกับ lot จริงของ order เดิม
- `api-contracts/17-stock-adjustments-contract.md`
- `api-contracts/18-stock-counts-contract.md`
- `api-contracts/19-product-returns-contract.md`

หมายเหตุ: ทั้งสามฟีเจอร์นี้เป็น "single-effect" operation (สร้างแล้วมีผลทันที ไม่มี state machine ต่อเนื่องแบบ approve/reject) จึงไม่มีเอกสาร lifecycle แยก ต่างจาก stock transfer ที่มีสถานะ pending/approved/rejected

## Part A: Oversell + Reconciliation (ส่วนขยายของ POS / Order)

### Business Rules

- order item หนึ่งบรรทัดสามารถส่ง `allowOversell: true` เพื่อขายเกิน stock ที่มีจริงได้ (เช่น กรณีร้านค้ายอมให้ค้างส่งลูกค้า)
- เมื่อ stock ของ lot ที่เลือกไม่พอ ระบบจะดึงเท่าที่มี แล้วบันทึกส่วนที่ขาดเป็น `oversoldQty` บน order item นั้น แทนที่จะ reject การขายทั้งบรรทัด
- ถ้าไม่ได้ส่ง `allowOversell` และ stock ไม่พอ ระบบต้อง reject การสร้าง order ตามพฤติกรรมเดิม
- `oversoldQty` ที่ค้างอยู่จะถูก "reconcile" (ลดยอดหนี้) โดยอัตโนมัติเมื่อ:
  - มี receive ใหม่ถูก import เข้า stock สำหรับสินค้าตัวเดียวกัน — reconcile กับ lot จริงที่เพิ่งสร้าง (FIFO ตามลำดับ order item ที่ค้าง)
  - มี stock adjustment ที่เป็นค่าบวก (`delta > 0`) สำหรับสินค้าตัวเดียวกัน — reconcile กับ marker สังเคราะห์ `ADJUST:<reason>` เพราะไม่มี lot จริงมารองรับ
- การ reconcile ไม่แตะยอด `ProductStock.quantity` ซ้ำ เพราะ `quantity` ถูกปรับไปแล้วตอน receive/adjustment เกิดขึ้นจริง — reconcile แค่ย้ายภาระจาก `oversoldQty` ไปเป็นรายการใน `stocks[]` ของ order item เพื่อเก็บ audit trail
- Invariant: `stock ณ เวลาใดๆ ≈ Σ lot.remaining − Σ oversoldQty` ต้องคงอยู่เสมอ

### Validation Rules

- `oversoldQty` ต้อง reconcile แบบ FIFO ตามลำดับ order item ที่เก่าที่สุดก่อน (เรียงตาม `_id` ของ order item)
- ห้าม reconcile เกินยอด `oversoldQty` ที่เหลืออยู่ หรือเกินยอด lot/adjustment ที่มีจริง

### Edge Cases

- สินค้าตัวเดียวกันมีหลาย order item ที่ oversold ค้างพร้อมกัน — ต้องไล่ reconcile ทีละรายการจนกว่า stock ใหม่จะหมดหรือหนี้หมด
- stock adjustment ที่เป็นค่าลบ (`delta < 0`) จะไม่ trigger การ reconcile ใดๆ

## Part B: Stock Adjustment

### Business Rules

- ทุก adjustment ต้องระบุเหตุผลจากชุดที่กำหนดไว้เท่านั้น: `นับสต็อก`, `ยาเสียหาย`, `ยาหมดอายุ`, `สูญหาย`, `อื่นๆ`
- `delta` เป็นบวกหรือลบก็ได้ แต่ต้องไม่เป็น 0
- ระบบต้องบันทึก `before`/`after`/`delta` ของ stock ที่ถูกปรับไว้เพื่อ audit
- adjustment ที่เป็นบวกเท่านั้นที่ trigger oversell reconciliation (ดู Part A)

### Validation Rules

- reason ต้องอยู่ในชุดที่กำหนด (ปฏิเสธถ้าไม่ตรง)
- stock ที่จะปรับต้องอยู่ใน branch เดียวกับผู้ทำรายการ
- adjustment ที่เป็นลบต้องไม่ทำให้ stock ติดลบ (reject ถ้า stock ไม่พอ)

### Edge Cases

- ปรับ stock ของ lot ที่ถูกใช้หมดไปแล้ว (`quantity = 0`) — ปรับเพิ่มได้ปกติ, ปรับลดต้อง reject

## Part C: Stock Count

### Business Rules

- การนับสต็อกคือการเทียบ `systemQuantity` (จาก `ProductStock.quantity` ปัจจุบัน) กับ `countedQuantity` (จำนวนที่นับได้จริง) ต่อ lot
- ถ้ามีผลต่าง (`delta != 0`) ระบบจะสร้าง stock adjustment ให้อัตโนมัติโดยใช้เหตุผล `นับสต็อก` — ไม่ต้องให้ผู้ใช้สร้าง adjustment แยกเอง
- ถ้าไม่มีผลต่างในบรรทัดนั้น ระบบจะไม่สร้าง adjustment ให้ (ไม่มี noise ใน audit log)
- เอกสารการนับสต็อก (`StockCount`) เก็บทุกบรรทัดที่นับ ไม่ว่าจะมีผลต่างหรือไม่ เพื่อเป็นหลักฐานว่านับครบทุกตัว

### Validation Rules

- ทุกบรรทัดต้องอ้างอิง `productId`/`stockId` ที่มีอยู่จริง
- การสร้าง adjustment ต่อบรรทัดใช้กฎเดียวกับ Part B ทั้งหมด (รวมถึง reconciliation)

### Edge Cases

- บรรทัดใดบรรทัดหนึ่งในชุดนับล้มเหลว (เช่น stock ไม่พบ) — รายการก่อนหน้าที่ apply ไปแล้วจะไม่ถูก rollback อัตโนมัติ (best-effort เหมือน stock transfer/receive ปัจจุบัน ไม่ใช่ transactional ทั้งชุด)

## Part D: Product Return

### Business Rules

- คืนสินค้าได้เฉพาะจำนวนที่ "ผูกกับ lot จริง" ของ order item เดิมเท่านั้น — ส่วนที่เป็น `oversoldQty` ที่ยังไม่ reconcile หรือส่วนที่ reconcile ด้วย synthetic marker (`ADJUST:<reason>`) ไม่สามารถคืนได้ เพราะไม่มี lot จริงให้คืนกลับ
- เพดานการคืนต่อ order item = `realLotQuantity(stocks) − returnedQty เดิม` (ไม่ใช่แค่ `quantity − returnedQty`)
- เมื่อคืนสำเร็จ ระบบจะคืน stock กลับไปยัง lot เดิมที่เคยถูกตัดออกไป โดยไล่ตามลำดับใน `stocks[]` ของ order item (ข้ามส่วนที่ถูกคืนไปแล้วจากการคืนครั้งก่อน และข้าม synthetic marker เสมอ)
- `returnedQty` บน order item จะถูกอัปเดตสะสมทุกครั้งที่คืนสำเร็จ

### Validation Rules

- ปฏิเสธถ้า `ReturnedQty + quantity ที่จะคืน > realLotQuantity`
- ปฏิเสธถ้า order item ที่อ้างอิงไม่ได้เป็นของ order ที่ระบุ
- ปฏิเสธถ้า order ไม่ได้อยู่ใน branch เดียวกับผู้ทำรายการ

### Edge Cases

- order item เดียวถูกคืนหลายครั้งบางส่วน (partial return) — ต้องไล่ allocate จาก lot ที่ถูกต้องทุกครั้งโดยไม่คืนซ้ำส่วนเดิม
- order item ที่มีทั้งส่วน lot จริงและส่วน oversold/synthetic ปนกัน — คืนได้เฉพาะสัดส่วน lot จริงเท่านั้น

## Expected Outcomes

- ร้านค้าที่ต้องการขายเกิน stock (pre-order/ค้างส่ง) ทำได้โดยไม่เสียความถูกต้องของ stock audit trail
- ยอด stock หลังปรับ/นับ/คืน ต้อง reconcile กับ order/receive ได้เสมอ ไม่มีส่วนที่ลอยหายไปจาก audit
- ผู้ดูแลสามารถตรวจสอบย้อนหลังได้ว่าทำไม stock ถึงเปลี่ยน (เหตุผล, ผู้ทำรายการ, เวลา) ในทุก mutation
