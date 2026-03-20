# Business Logic Documentation

เอกสารชุดนี้อธิบาย business logic ของระบบ Smart Pharmacy Web POS ในระดับ feature โดยเน้น:

- วัตถุประสงค์ทางธุรกิจ
- ผู้มีส่วนเกี่ยวข้อง
- ข้อมูลนำเข้าและผลลัพธ์
- กฎการทำงานของระบบ
- เงื่อนไขการตรวจสอบและข้อจำกัด
- กรณีพิเศษและข้อควรระวัง
- พฤติกรรมฝั่ง Web/Frontend และ user workflow

## เอกสารในชุดนี้

- `01-auth-and-access-control.md`
- `02-product-and-inventory.md`
- `03-pos-and-clinical-support.md`
- `04-regulatory-reports.md`
- `05-patient-and-crm.md`
- `06-dashboard-and-analytics.md`
- `07-settings-and-master-data.md`
- `08-promotions-and-stock-transfer.md`

## เอกสาร Workflow / Sequence Flow

- `flows/README.md`
- `flows/01-goods-receipt-flow.md`
- `flows/02-pos-order-flow.md`
- `flows/03-report-generation-flow.md`
- `flows/04-patient-customer-flow.md`

## เอกสาร Lifecycle / State Transition

- `lifecycle/README.md`
- `lifecycle/01-order-lifecycle.md`
- `lifecycle/02-receive-lifecycle.md`
- `lifecycle/03-promotion-lifecycle.md`
- `lifecycle/04-stock-transfer-lifecycle.md`

## เอกสาร Web Screens

- `screens/README.md`
- `screens/01-dashboard-screen.md`
- `screens/02-products-screen.md`
- `screens/03-receives-screen.md`
- `screens/04-reports-screen.md`
- `screens/05-pos-screen.md`
- `screens/06-customers-screen.md`
- `screens/07-patients-screen.md`
- `screens/08-settings-screen.md`
- `screens/09-promotions-screen.md`
- `screens/10-stock-transfers-screen.md`

## เอกสาร API Contracts

- `api-contracts/README.md`
- `api-contracts/01-auth-and-context-contract.md`
- `api-contracts/02-products-contract.md`
- `api-contracts/03-receives-contract.md`
- `api-contracts/04-pos-orders-contract.md`
- `api-contracts/05-reports-contract.md`
- `api-contracts/06-patients-and-customers-contract.md`
- `api-contracts/07-settings-contract.md`

## หลักการออกแบบ business logic

- ระบบต้องรักษาความถูกต้องของข้อมูลก่อนความสะดวก
- การเปลี่ยนแปลงที่กระทบสต็อกต้องตรวจสอบย้อนหลังได้
- รายการขายยาควบคุมต้องตรวจสอบผู้เกี่ยวข้องและข้อมูลกำกับทุกครั้ง
- รายงานที่เกี่ยวข้องกับกฎหมายต้องออกรายงานจากข้อมูลจริงที่ audit ได้
- ฝั่ง Backend รับผิดชอบด้าน data integrity, authorization, persistence และ data endpoints
- ฝั่ง Frontend รับผิดชอบ workflow การใช้งาน, validation เชิงประสบการณ์ผู้ใช้, และการแสดงผลเอกสาร
- เอกสารแต่ละไฟล์ควรอ่านได้ทั้งในมุม domain logic และมุม user interaction บนหน้าเว็บ
