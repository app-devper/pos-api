# 06. Dashboard and Analytics

## วัตถุประสงค์ทางธุรกิจ

ให้ผู้บริหารและผู้ปฏิบัติงานมองเห็นภาพรวมของยอดขาย สต็อก และความเสี่ยงเชิงปฏิบัติการในเวลาที่รวดเร็วพอสำหรับการตัดสินใจ

## Workflow References

- `flows/13-dashboard-analytics-flow.md` — dashboard query flow ระดับภาพรวม
- `lifecycle/16-dashboard-query-lifecycle.md` — lifecycle ของ derived dashboard data
- `api-contracts/16-dashboard-contract.md` — endpoint contracts
- `screens/01-dashboard-screen.md` — พฤติกรรมการแสดงผลบนหน้า dashboard

## End-to-End Workflow Summary

### 1. Open Dashboard with Branch Context

- ผู้ใช้เปิด dashboard ภายใต้ branch ปัจจุบัน
- Frontend เรียก summary, charts และ alert widgets แยกกันตามประเภทข้อมูล
- Backend validate query params และ branch context ก่อนเริ่ม aggregate

### 2. Compute Sales Analytics

- summary, daily chart และ monthly chart ใช้ order data ของ branch ปัจจุบัน
- ระบบนับเฉพาะ order ที่อยู่ในสถานะที่ถือเป็นรายได้จริง
- revenue, cost และ profit ต้องใช้ semantics เดียวกับรายงาน export

### 3. Compute Inventory Analytics

- low stock, stock report, dead stock และ expiring ดึงจาก inventory data ที่ scope ตาม branch
- metric ที่เกี่ยวกับ stock ต้องไม่ปนข้าม unit โดยผิดความหมาย
- dead stock และ expiring ต้องสะท้อน stock ที่ใช้งานจริงของสาขาปัจจุบัน

### 4. Render Derived Views

- Frontend นำผลลัพธ์แต่ละ endpoint ไป render เป็น cards, charts, tables และ alerts
- ถ้าบาง widget ไม่มีข้อมูล ต้องแสดง empty state ชัดเจน
- ถ้า widget ใด fail ต้องไม่ทำให้ semantics ของ widget อื่นเพี้ยนตาม

### 5. Cross-Check with Reports

- dashboard ต้องให้ตัวเลขสอดคล้องกับ report/export ที่อิงกฎเดียวกัน
- ถ้ามีความต่างระหว่าง dashboard กับ report ต้องถือเป็น bug ทางธุรกิจ ไม่ใช่ความต่างด้าน UI

## Business Rules

### 1. Sales Summary

- ยอดขายต้องอิงจาก order ที่สมบูรณ์และสถานะที่นับเป็นรายได้จริง
- dashboard ไม่ควรรวม order ที่ถูกยกเลิกหรือ voided

### 2. Cost and Profit

- ต้นทุนและกำไรต้องคำนวณจากข้อมูลต้นทุนสินค้าจริงในระบบ
- หากไม่มีต้นทุนของบางรายการ ต้องกำหนดวิธีตีความให้ชัด ไม่เช่นนั้นกำไรอาจเพี้ยน

### 3. Inventory Monitoring

- low stock ดูจาก min stock เทียบกับยอดคงเหลือจริง
- expiring items ดูจากวันหมดอายุใน lot
- dead stock ดูจากการไม่เคลื่อนไหวเกินช่วงเวลาที่กำหนด

### 4. ABC Analysis

- ต้องคำนวณจากยอดขายสะสมในช่วงเวลาที่กำหนด
- การจัดกลุ่ม A/B/C ใช้เพื่อวางแผนสต็อกและการสั่งซื้อ ไม่ใช่เพียงการแสดงผลสวยงาม

## Validation Rules

- date range ของ dashboard ต้องถูกต้อง
- ข้อมูลสรุปต้องผูกกับ branch context
- metric แต่ละตัวต้องอิงชุดข้อมูลที่นิยามไว้ชัดเจน

## Edge Cases

- หากข้อมูลย้อนหลังขาดบางวัน กราฟต้องยัง render ได้โดยไม่ทำให้ผู้ใช้เข้าใจผิด
- ถ้าไม่มีข้อมูลในช่วงที่เลือก ต้องแสดง empty state อย่างชัดเจน

## Expected Outcomes

- ผู้ใช้เห็นภาพรวมธุรกิจเร็ว
- ระบุปัญหา stock, expiry, และยอดขายตกได้เร็ว
- ใช้ข้อมูลชุดเดียวกับระบบปฏิบัติการจริงเพื่อลดความคลาดเคลื่อน
