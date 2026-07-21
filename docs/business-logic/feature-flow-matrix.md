# Feature Flow Matrix

ตารางนี้ใช้สรุปว่าแต่ละ feature มีเอกสารรองรับครบในมิติหลักแค่ไหน:

- `Domain` = เอกสาร business/domain overview
- `Flow` = sequence flow / workflow
- `Lifecycle` = state transition / lifecycle
- `Contract` = API contract
- `Screen` = screen / frontend behavior

## Matrix

| Feature | Domain | Flow | Lifecycle | Contract | Screen | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Auth / Access Control | `01-auth-and-access-control.md` | - | - | `api-contracts/01-auth-and-context-contract.md` | - | เน้น middleware และ request context |
| Product / Inventory | `02-product-and-inventory.md` | `flows/11-product-management-flow.md` | `lifecycle/05-product-lifecycle.md` | `api-contracts/02-products-contract.md` | `screens/02-products-screen.md` | ซับซ้อนที่สุด และเชื่อมกับ receive/order/report |
| Receive | `02-product-and-inventory.md` | `flows/01-goods-receipt-flow.md` | `lifecycle/02-receive-lifecycle.md` | `api-contracts/03-receives-contract.md` | `screens/03-receives-screen.md` | ครอบตั้งแต่เอกสารรับเข้าไปจน import เข้า stock |
| POS / Order | `03-pos-and-clinical-support.md` | `flows/02-pos-order-flow.md` | `lifecycle/01-order-lifecycle.md` | `api-contracts/04-pos-orders-contract.md` | `screens/05-pos-screen.md` | รวม order item และ payment behavior |
| Reports / Documents | `04-regulatory-reports.md` | `flows/03-report-generation-flow.md`, `flows/10-khy-sale-flow.md` | `lifecycle/08-report-document-lifecycle.md` | `api-contracts/05-reports-contract.md` | `screens/04-reports-screen.md` | ครอบ PDF/Excel/KHY exports |
| Patients | `05-patient-and-crm.md` | `flows/06-patient-management-flow.md` | `lifecycle/06-patient-lifecycle.md` | `api-contracts/06-patients-and-customers-contract.md` | `screens/07-patients-screen.md` | เชื่อมกับ allergy check และประวัติผู้ป่วย |
| Customers | `05-patient-and-crm.md` | `flows/05-customer-management-flow.md` | `lifecycle/07-customer-lifecycle.md` | `api-contracts/06-patients-and-customers-contract.md` | `screens/06-customers-screen.md` | customer model เป็น global มากกว่า entity อื่น |
| Customer History | `05-patient-and-crm.md` | `flows/18-customer-history-flow.md` | `lifecycle/15-customer-history-lifecycle.md` | `api-contracts/15-customer-histories-contract.md` | `screens/06-customers-screen.md` | ใช้ branch-scoped CRM activity log |
| Dashboard / Analytics | `06-dashboard-and-analytics.md` | `flows/13-dashboard-analytics-flow.md` | `lifecycle/16-dashboard-query-lifecycle.md` | `api-contracts/16-dashboard-contract.md` | `screens/01-dashboard-screen.md` | semantics ต้องตรงกับ report/export |
| Settings | `07-settings-and-master-data.md` | `flows/07-settings-management-flow.md` | `lifecycle/13-settings-lifecycle.md` | `api-contracts/07-settings-contract.md` | `screens/08-settings-screen.md` | เป็น branch-level config; feature toggles ยังเป็น frontend-managed |
| Promotions | `08-promotions-and-stock-transfer.md` | `flows/08-promotion-management-flow.md` | `lifecycle/03-promotion-lifecycle.md` | `api-contracts/08-promotions-contract.md` | `screens/09-promotions-screen.md` | มี apply flow แยกจาก CRUD |
| Stock Transfer | `08-promotions-and-stock-transfer.md` | `flows/09-stock-transfer-flow.md` | `lifecycle/04-stock-transfer-lifecycle.md` | `api-contracts/09-stock-transfers-contract.md` | `screens/10-stock-transfers-screen.md` | มี create / approve / reject semantics |
| Branches | `07-settings-and-master-data.md` | `flows/14-branch-management-flow.md` | `lifecycle/09-branch-lifecycle.md` | `api-contracts/10-branches-contract.md` | - | เป็น boundary หลักของ branch-scoped data |
| Employees | `07-settings-and-master-data.md` | `flows/15-employee-management-flow.md` | `lifecycle/10-employee-lifecycle.md` | `api-contracts/11-employees-contract.md` | - | เชื่อม role + branch mapping |
| Suppliers | `07-settings-and-master-data.md` | `flows/16-supplier-management-flow.md` | `lifecycle/11-supplier-lifecycle.md` | `api-contracts/12-suppliers-contract.md` | - | ใช้กับ receives และ purchase reports |
| Categories | `07-settings-and-master-data.md` | `flows/17-category-management-flow.md` | `lifecycle/12-category-lifecycle.md` | `api-contracts/13-categories-contract.md` | - | ใช้ใน product grouping/filtering |
| Stock Adjustment | `09-stock-adjustments-counts-and-returns.md` | `flows/19-stock-adjustment-flow.md` | - | `api-contracts/17-stock-adjustments-contract.md` | - | single-effect operation ไม่มี state machine จึงไม่มี lifecycle doc; trigger oversell reconciliation เมื่อ delta บวก |
| Stock Count | `09-stock-adjustments-counts-and-returns.md` | `flows/20-stock-count-flow.md` | - | `api-contracts/18-stock-counts-contract.md` | - | เรียกใช้ stock adjustment logic ภายในต่อบรรทัดที่มีผลต่าง |
| Product Return | `09-stock-adjustments-counts-and-returns.md` | `flows/21-product-return-flow.md` | - | `api-contracts/19-product-returns-contract.md` | - | คืนได้เฉพาะจำนวนที่ผูกกับ lot จริงของ order เดิม ไม่รวม oversold/synthetic adjustment |

## Coverage Notes

- feature ที่มีครบเกือบทุกมิติมากที่สุดคือ `product`, `receive`, `order`, `reports`, `patients`, `promotions`, และ `stock transfer`
- feature ที่ยังไม่มี screen doc แยกชัดคือ `branches`, `employees`, `suppliers`, `categories`, `stock adjustment`, `stock count`, และ `product return` (สามตัวหลังยังไม่มี dedicated screen ในโค้ด frontend)
- `stock adjustment`, `stock count`, และ `product return` ไม่มี lifecycle doc โดยตั้งใจ เพราะเป็น single-effect operation (สร้างแล้วมีผลทันที) ไม่ใช่ multi-state workflow แบบ stock transfer
- feature toggles ใน settings ปัจจุบันเป็น stored configuration ที่ frontend/consumer ใช้ตัดสินใจเอง ยังไม่ใช่ backend runtime gate ของทุก API
- ถ้าจะเพิ่มเอกสารต่อ รอบถัดไปควรเลือกจากมิติที่ยังว่างอยู่ก่อน ไม่ใช่เพิ่ม flow ซ้ำกับ feature ที่ coverage ดีอยู่แล้ว
