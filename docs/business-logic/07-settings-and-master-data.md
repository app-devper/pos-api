# 07. Settings and Master Data

## วัตถุประสงค์ทางธุรกิจ

ทำให้ระบบสามารถปรับพฤติกรรมได้ตามสาขาและนโยบาย โดยไม่ต้องแก้โค้ดทุกครั้ง และทำให้ข้อมูลหลักที่ใช้ร่วมกันทั้งระบบมีความสอดคล้องกัน

## Areas in Scope

- System Settings
- Feature Toggles
- Branches
- Employees
- Categories
- Customers
- Suppliers

## Workflow References

- `flows/07-settings-management-flow.md` — branch settings management
- `flows/14-branch-management-flow.md` — branch management
- `flows/15-employee-management-flow.md` — employee management
- `flows/16-supplier-management-flow.md` — supplier management
- `flows/17-category-management-flow.md` — category management
- `lifecycle/09-branch-lifecycle.md` — branch lifecycle
- `lifecycle/10-employee-lifecycle.md` — employee lifecycle
- `lifecycle/11-supplier-lifecycle.md` — supplier lifecycle
- `lifecycle/12-category-lifecycle.md` — category lifecycle
- `lifecycle/13-settings-lifecycle.md` — settings lifecycle

## End-to-End Workflow Summary

### 1. Establish Branch Boundary

- ระบบสร้างและดูแล branch master ก่อน
- branch เป็น boundary หลักของการมองเห็นข้อมูล, inventory scope, dashboard, reports และ settings

### 2. Map Employees to Branch and Role

- employee ถูกผูกกับ branch และ role
- middleware ใช้ข้อมูลนี้ในการตีความ branch context และ authorization ทุก feature

### 3. Maintain Branch-Level Settings

- ผู้ดูแลอัปเดต company info, receipt footer, PromptPay ID และค่า config อื่นตามสาขา
- report/document paths และบาง behavior ของระบบอ้าง settings ของ branch ปัจจุบัน

### 4. Maintain Shared Master Data

- categories ใช้กับการจัดกลุ่มสินค้าและ filtering
- suppliers ใช้กับ receive และ purchase reporting
- customers ใช้กับ purchase history, CRM และ patient/customer linkage

### 5. Propagate to Operational Features

- receive ใช้ supplier และ branch
- product ใช้ category และ branch
- order/report/dashboard ใช้ branch, employee context และ settings
- CRM / patient flows ใช้ customer master ประกอบกับ patient domain

### 6. Preserve Historical Meaning

- master data ที่ถูกแก้หรือปิดใช้งานต้องไม่ทำให้เอกสารย้อนหลังสูญเสียความหมาย
- แนวทางที่ปลอดภัยกว่าคือ inactive / status-based control มากกว่าลบ hard delete โดยไม่จำเป็น

## Business Rules

### 1. System Settings

- setting เป็นแหล่งข้อมูลกลางของชื่อบริษัท ที่อยู่ เบอร์โทร ข้อความท้ายใบเสร็จ และค่าพฤติกรรมระบบบางส่วน
- setting ระดับสาขาต้องถูกใช้กับเอกสารและรายงานของสาขานั้น
- หากไม่มีค่า setting บางตัว ระบบควร fallback เป็นค่า default ที่ปลอดภัย

### 2. Feature Toggles

- ใช้เปิด/ปิดฟีเจอร์เชิงธุรกิจ เช่น patient feature
- ปัจจุบันค่ากลุ่มนี้ถูกเก็บเป็น branch-level config และถูกคาดหวังให้ frontend/consumer ใช้ตัดสินใจซ่อนหรือเปิด flow
- หากต้องการให้ backend ปฏิเสธการเข้าถึง logic ที่ไม่ควรถูกใช้ ต้องมี runtime gate เพิ่มในชั้น API/middleware/usecase

### 3. Branches and Employees

- พนักงานต้องสังกัดสาขาที่ชัดเจน
- branch context ถูกใช้เป็นตัวกรองหลักของข้อมูลธุรกิจ
- การย้ายสังกัดพนักงานมีผลต่อสิทธิ์การมองเห็นข้อมูล

### 4. Categories, Suppliers, Customers

- เป็น master data ที่ต้องใช้ร่วมกันในหลาย workflow
- การลบข้อมูลหลักไม่ควรทำให้เอกสารประวัติหรือรายการย้อนหลังเสียความสมบูรณ์
- หากต้องปิดใช้งาน ควรใช้แนวทาง inactive มากกว่าลบทิ้ง

## Validation Rules

- ชื่อ field หลักต้องไม่ว่างเมื่อเป็นข้อมูลบังคับ
- branch / employee relation ต้องอ้างอิงถึงกันได้จริง
- feature flag ต้องอ่านค่าได้แน่นอน ไม่คลุมเครือ

## Expected Outcomes

- ระบบปรับค่าใช้งานได้ง่าย
- เอกสารและรายงานแสดงข้อมูลบริษัท/สาขาถูกต้อง
- master data รองรับ workflow อื่นโดยไม่ทำให้ referential meaning สูญหาย
