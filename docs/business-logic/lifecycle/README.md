# Entity Lifecycle and State Transition Documentation

เอกสารชุดนี้อธิบายการเปลี่ยนสถานะของ entity หลักในระบบ พร้อมกฎทางธุรกิจ ผลกระทบต่อข้อมูล และพฤติกรรมที่ Frontend ควรสะท้อนให้ผู้ใช้เห็น

## เอกสารในชุดนี้

- `01-order-lifecycle.md`
- `02-receive-lifecycle.md`
- `03-promotion-lifecycle.md`
- `04-stock-transfer-lifecycle.md`
- `05-product-lifecycle.md`
- `06-patient-lifecycle.md`
- `07-customer-lifecycle.md`
- `08-report-document-lifecycle.md`
- `09-branch-lifecycle.md`
- `10-employee-lifecycle.md`
- `11-supplier-lifecycle.md`
- `12-category-lifecycle.md`
- `13-settings-lifecycle.md`
- `15-customer-history-lifecycle.md`
- `16-dashboard-query-lifecycle.md`

## วิธีใช้เอกสารชุดนี้

แต่ละไฟล์อธิบาย:

- สถานะหลักของ entity
- เหตุการณ์ที่ทำให้เปลี่ยนสถานะ
- เงื่อนไขก่อนและหลังการเปลี่ยนสถานะ
- ผลกระทบต่อข้อมูลอื่น
- พฤติกรรมที่หน้าเว็บควรแสดง
- ข้อผิดพลาดที่ต้องป้องกัน
