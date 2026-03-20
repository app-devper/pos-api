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

## Business Rules

### 1. System Settings

- setting เป็นแหล่งข้อมูลกลางของชื่อบริษัท ที่อยู่ เบอร์โทร ข้อความท้ายใบเสร็จ และค่าพฤติกรรมระบบบางส่วน
- setting ระดับสาขาต้องถูกใช้กับเอกสารและรายงานของสาขานั้น
- หากไม่มีค่า setting บางตัว ระบบควร fallback เป็นค่า default ที่ปลอดภัย

### 2. Feature Toggles

- ใช้เปิด/ปิดฟีเจอร์เชิงธุรกิจ เช่น patient feature
- การปิด feature ต้องซ่อนทั้ง UI และปฏิเสธการเข้าถึง logic ที่ไม่ควรถูกใช้

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
