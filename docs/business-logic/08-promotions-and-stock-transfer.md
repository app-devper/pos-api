# 08. Promotions and Stock Transfer

## วัตถุประสงค์ทางธุรกิจ

รองรับการทำการตลาดผ่านโปรโมชัน และรองรับการกระจายสินค้าในหลายสาขาโดยไม่ทำให้ stock integrity เสียหาย

## Workflow References

- `flows/08-promotion-management-flow.md` — promotion CRUD และ apply behavior
- `flows/09-stock-transfer-flow.md` — stock transfer flow
- `lifecycle/03-promotion-lifecycle.md` — promotion lifecycle
- `lifecycle/04-stock-transfer-lifecycle.md` — stock transfer lifecycle
- `api-contracts/08-promotions-contract.md` — promotion contracts
- `api-contracts/09-stock-transfers-contract.md` — stock transfer contracts
- `screens/09-promotions-screen.md` — promotion UI behavior
- `screens/10-stock-transfers-screen.md` — stock transfer UI behavior

## End-to-End Workflow Summary

### 1. Define Commercial Rules

- ผู้ดูแลสร้าง promotion พร้อม code, type, date range, min purchase และ product conditions
- ผู้ดูแลสร้าง stock transfer request พร้อม from branch, to branch และรายการสินค้า

### 2. Validate Before Use

- promotion ต้องอยู่ในช่วงเวลาที่ใช้งานได้และมีสถานะพร้อมใช้
- stock transfer ต้องมีต้นทาง/ปลายทางต่างกัน, quantity ถูกต้อง และสินค้าอ้างอิงได้จริง

### 3. Apply at Runtime

- promotion ถูก apply ระหว่าง order flow ก่อนยืนยันการขาย
- stock transfer ถูกสร้างเป็นคำขอและ reserve/ตรวจ stock ต้นทางตามกติกาของระบบ

### 4. Commit Business Effect

- promotion เปลี่ยนยอดสุทธิของ order แต่ต้องไม่ทำให้ยอดติดลบ
- stock transfer เมื่อ approve จึงย้าย stock ระหว่างสาขาและอัปเดตสถานะ
- reject หรือ failure ระหว่างทางต้องไม่ทำให้ stock ค้างครึ่งทาง

### 5. Preserve Audit and Operational Consistency

- promotion ที่ถูก apply ต้องย้อนดูได้จาก order data
- stock transfer ต้องย้อนดูได้จาก transfer document, stock movement และ status history
- ทั้งสอง feature ต้องไม่ทำให้ dashboard/report ให้ค่าคลาดเคลื่อนจากข้อมูลจริง

## Part A: Promotions

### Business Rules

- โปรโมชันมีช่วงเวลาเริ่มต้นและสิ้นสุดชัดเจน
- โปรโมชันต้องอยู่ในสถานะ `ACTIVE` จึงใช้ได้
- โปรโมชันแบบเปอร์เซ็นต์ต้องมีเพดานส่วนลดถ้าธุรกิจกำหนด
- โปรโมชันแบบกำหนดสินค้าเฉพาะใช้ได้เฉพาะสินค้าในเงื่อนไข
- โปรโมชันต้องไม่ทำให้ยอดขายสุทธิติดลบ

### Validation Rules

- code ต้องไม่ซ้ำภายในขอบเขตที่กำหนด
- value ต้องมากกว่า 0
- startDate ต้องไม่มากกว่า endDate
- minPurchase ต้องไม่ติดลบ

### Edge Cases

- ลูกค้ามีหลายโปรโมชันพร้อมกัน ต้องกำหนด rule ว่าซ้อนกันได้หรือไม่
- สินค้าบางตัวถูกยกเว้น ต้องถูกกรองออกจากการคำนวณ

## Part B: Stock Transfer

### Business Rules

- การโอนสต็อกเริ่มจากคำขอที่มีต้นทางและปลายทางชัดเจน
- ทุกขั้นเป็นการเคลื่อนไหวของ stock ผ่าน Stock ledger (ADR-0001) และเขียน product history ของตัวเอง:
  - **สร้างคำขอ** — กัน (ตัด) จำนวนออกจาก stock ต้นทางทันที เพื่อไม่ให้ของที่กำลังจะโอนถูกขาย; stock ต้นทางต้องพอ
  - **อนุมัติ** — สร้าง stock ที่สาขาปลายทาง (lot, ต้นทุน, ราคา, วันหมดอายุตามต้นทาง) และ settle oversell ของปลายทางในหน่วยเดียวกันก่อน (ADR-0002)
  - **ปฏิเสธ** — คืนจำนวนที่กันไว้เข้า stock ต้นทาง (และ settle oversell ของต้นทางในหน่วยเดียวกัน)
- stock ไม่ติดลบในทุกขั้น

### Validation Rules

- ไม่อนุญาตให้โอนไปสาขาเดียวกัน
- quantity ต้องมากกว่า 0
- product และ unit ต้องอ้างอิงได้จริง

### Edge Cases

- stock ต้นทางถูกกันไว้ตั้งแต่สร้างคำขอ การขายระหว่างรออนุมัติจึงไม่แย่งของที่กำลังโอน
- การอนุมัติหรือปฏิเสธซ้ำ (คำขอที่ไม่ใช่ PENDING แล้ว) ต้องถูกปฏิเสธ ไม่ทำให้ stock ถูกย้ายซ้ำ

## Expected Outcomes

- โปรโมชันถูกนำไปใช้ได้อย่างคาดการณ์ได้
- การโอนสต็อกไม่ทำให้ยอดคงเหลือเพี้ยน
- ทุกการเปลี่ยนแปลงสำคัญสามารถตรวจสอบย้อนหลังได้
