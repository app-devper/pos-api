# 02. Product and Inventory

## วัตถุประสงค์ทางธุรกิจ

ควบคุมข้อมูลสินค้า ยา Lot วันหมดอายุ หน่วยนับ และยอดคงเหลือให้ถูกต้อง เพื่อรองรับการขาย การรับเข้า และการตรวจสอบย้อนหลัง

## ผู้มีส่วนเกี่ยวข้อง

- ผู้ดูแลสินค้า
- พนักงานรับสินค้า
- พนักงานขาย
- ผู้บริหารสาขา

## Domain Objects

- Product
- Product Unit
- Product Price
- Product Stock
- Product Lot
- Receive / Receive Item
- Product History

## Workflow References

- `flows/11-product-management-flow.md` — product master, unit, price, stock, lot, history
- `flows/01-goods-receipt-flow.md` — receive/import เข้าสต็อก
- `flows/09-stock-transfer-flow.md` — โอนสต็อกระหว่างสาขา
- `lifecycle/05-product-lifecycle.md` — สถานะและการเปลี่ยนผ่านของข้อมูลสินค้า

## End-to-End Workflow Summary

### 1. Create Product Master

- ผู้ดูแลสร้างสินค้าใหม่ด้วยข้อมูลหลัก เช่น serial number, ชื่อสินค้า, default unit, cost price, selling price และข้อมูลยา
- Backend สร้าง product master พร้อม default unit และ default price
- ระบบสร้าง product history เพื่อให้การสร้างสินค้า audit ย้อนหลังได้

### 2. Expand Commercial Structure

- ผู้ใช้สามารถเพิ่ม unit เพิ่มเติม เช่น กล่อง / แผง / เม็ด
- ผู้ใช้สามารถเพิ่มหรือแก้ price ต่อหน่วย
- ทุกการแก้ unit / price ต้องไม่ทำให้รายการรับเข้าและขายย้อนหลังเสียความหมาย

### 3. Add or Adjust Stock

- stock ถูกสร้างจาก receive, product receive, stock adjustment หรือ stock transfer
- stock ทุกก้อนต้องผูกกับ branch, unit, quantity, และข้อมูล lot/expiry เมื่อเกี่ยวข้อง
- การเปลี่ยนแปลง stock ต้องสร้าง product history และใช้ branch-scoped balance เท่านั้น

### 4. Receive and Lot Tracking

- เมื่อรับสินค้า ระบบสร้าง receive document และ receive items
- เมื่อ import เข้า stock ระบบเพิ่ม quantity, lot, expire date, total cost และ history ให้สอดคล้องกัน
- lot และ expire notify ต้องถูกมองในบริบท branch เดียวกัน

### 5. Sell and Deduct Inventory

- order flow ใช้ unit และ price ของสินค้าตาม transaction จริง
- ระบบตัด stock หรือ sold-first ตามกติกาของสินค้า
- order item, stock movement และ product history ต้องคงความสอดคล้องกันในระดับ transaction

### 6. Transfer Between Branches

- stock transfer reserve stock ต้นทางก่อน
- เมื่อ approve จึงเพิ่ม stock ปลายทางและอัปเดตสถานะ
- การ reject หรือ fail ระหว่างทางต้องไม่ทำให้ stock ข้ามสาขาค้างครึ่งทาง

### 7. Report and Audit

- dashboard, stock report, low stock, dead stock, expiring และ KHY reports ใช้ข้อมูล product/inventory เป็นฐาน
- รายงานต้องไม่ปนข้าม branch หรือข้าม unit โดยผิดความหมาย
- product history เป็นแหล่งตรวจสอบย้อนหลังสำคัญที่สุดของ feature นี้

## Business Rules

### 1. Product Master

- สินค้าหนึ่งรายการต้องมีข้อมูลหลักอย่างน้อย ชื่อ หน่วย ราคาทุน ราคาขาย และสถานะ
- สินค้าสามารถมีข้อมูลด้านยาเพิ่มเติม เช่น ประเภทยา ชื่อสามัญ ข้อห้ามใช้ และผลข้างเคียง
- สินค้าที่ถูกปิดใช้งานไม่ควรถูกนำไปขายใหม่ แต่ข้อมูลต้องยังค้นย้อนหลังได้

### 2. Drug Classification

- สินค้าสามารถระบุประเภทตามกฎหมายยาได้
- ประเภทของยามีผลต่อ workflow การขาย การบันทึกข้อมูล และรายงาน ข.ย.
- การจัดหมวดผิดมีผลทางกฎหมาย จึงต้องแก้ไขได้เฉพาะผู้มีสิทธิ์

### 3. Multi-Unit Logic

- สินค้าหนึ่งตัวมีได้หลายหน่วย เช่น กล่อง แผง เม็ด
- ระบบต้องมีหน่วยฐานสำหรับคำนวณสต็อกจริง
- ทุกการรับเข้าและการขายในหน่วยรองต้องแปลงกลับเป็นหน่วยฐานก่อนตัดสต็อก
- ราคาทุนและราคาขายอาจต่างกันตามแต่ละหน่วยได้

### 4. Barcode Mapping

- บาร์โค้ดหนึ่งชุดต้องชี้ไปยังสินค้าและหน่วยที่แน่นอน
- การสแกนบาร์โค้ดต้องเลือกหน่วยนั้นให้อัตโนมัติ
- หากบาร์โค้ดซ้ำกันในระบบ ต้องถือเป็นข้อมูลผิดพลาดเชิงธุรกิจ

### 5. Goods Receipt

- การรับสินค้า 1 ครั้งสร้างเอกสารรับสินค้า 1 ใบ
- ในเอกสารรับสินค้าหนึ่งใบมีได้หลายรายการ
- แต่ละรายการควรมีอย่างน้อย product, quantity, costPrice, lotNumber, expireDate
- การบันทึกรับสินค้าเสร็จสมบูรณ์ต้องอัปเดต stock, lot, และ product history ให้สอดคล้องกัน

### 6. Lot and Expiry

- สินค้าที่ต้องติดตาม lot ต้องเก็บ lot number และวันหมดอายุ
- สินค้าที่ใกล้หมดอายุควรถูกแจ้งเตือนล่วงหน้า
- หาก POS UI ใช้นโยบาย FEFO ตอนเลือก lot ระบบควรส่ง `item.Stocks` ที่สะท้อนลำดับนั้นมาให้ backend ใช้ตัด stock

### 7. Product History

- ทุกการเคลื่อนไหวที่กระทบสต็อกควรสร้างประวัติสินค้า
- ประวัติต้องตอบได้ว่าเกิดจาก receive, order, adjustment หรือ transfer
- Product history เป็นแหล่งอ้างอิงสำคัญในการตรวจสอบย้อนหลัง

## Validation Rules

- quantity ต้องมากกว่า 0 เมื่อรับเข้า
- costPrice และ price ต้องไม่ติดลบ
- expireDate ต้องสมเหตุสมผลเมื่อสินค้าอยู่ในกลุ่มที่ต้องติดตาม lot
- การแปลงหน่วยต้องไม่ทำให้ base quantity ผิด
- สต็อกหลังคำนวณต้องไม่ติดลบใน workflow ปกติ

## Edge Cases

- รับสินค้าเข้าซ้ำ lot เดิม ต้องกำหนดนโยบายว่ารวม lot เดิมหรือเพิ่ม record ใหม่
- หากสินค้ามีหลายหน่วยแต่ unit mapping ไม่ครบ ต้องไม่ให้ทำรายการต่อ
- หาก product ถูกปิดใช้งาน แต่ยังมี stock คงเหลือ ต้องยังดูข้อมูลย้อนหลังได้

## Expected Outcomes

- สต็อกคงเหลือและ lot balance ถูกต้อง
- ข้อมูลพร้อมใช้กับ POS, dashboard, และรายงานตามกฎหมาย
- การตรวจสอบย้อนหลังทำได้จากเอกสาร receive และ product history
