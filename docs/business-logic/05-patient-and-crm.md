# 05. Patient and CRM

## วัตถุประสงค์ทางธุรกิจ

จัดเก็บข้อมูลผู้ป่วยและลูกค้าเพื่อใช้ในการดูแลต่อเนื่อง ลดความเสี่ยงจากการแพ้ยา และเพิ่มประสิทธิภาพในการให้บริการหน้าร้าน

## Domain Areas

- Patient Profile
- Allergy
- Chronic Disease
- Customer Profile
- Customer Purchase History

## Workflow References

- `flows/04-patient-customer-flow.md` — ความสัมพันธ์ระดับภาพรวมระหว่าง patient กับ customer
- `flows/05-customer-management-flow.md` — customer CRUD และการใช้งานฝั่ง CRM
- `flows/06-patient-management-flow.md` — patient management
- `flows/18-customer-history-flow.md` — customer activity log
- `lifecycle/06-patient-lifecycle.md` — patient lifecycle
- `lifecycle/07-customer-lifecycle.md` — customer lifecycle
- `lifecycle/15-customer-history-lifecycle.md` — customer history lifecycle

## End-to-End Workflow Summary

### 1. Create or Identify Customer Context

- พนักงานสามารถสร้างหรือค้นหาลูกค้าเพื่อผูกกับการขาย
- customer context ใช้กับ purchase history, CRM activity และ follow-up เชิงธุรกิจ

### 2. Create or Identify Patient Context

- ถ้าการขายเกี่ยวข้องกับการดูแลทางคลินิก ผู้ใช้เลือกหรือสร้าง patient profile
- patient domain แยกจาก customer แม้ในเชิงธุรกิจจะอาจเป็นคนเดียวกัน

### 3. Maintain Clinical Information

- ระบบเก็บ allergy, chronic disease และข้อมูลสุขภาพที่เกี่ยวข้อง
- ข้อมูลกลุ่มนี้ถูกใช้ใน checkpoint ก่อนการขายและการจ่ายยา

### 4. Use Patient Data During Sale

- เมื่อมี patient context ระบบตรวจ allergy check ก่อน commit order
- warning หรือ block ที่สำคัญต้องเกิดก่อนตัด stock และปิดการขาย

### 5. Record Clinical Follow-up

- หลังการขายที่เกี่ยวข้องกับผู้ป่วย ระบบเก็บ patient context และ order history ไว้ใช้ติดตามต่อ
- ประวัติการขายและข้อมูลผู้ป่วยใช้กับ patient review และการติดตามการรักษา

### 6. Build CRM and Service History

- customer history ถูกเพิ่มจากกิจกรรมที่เกี่ยวข้องกับลูกค้าในสาขานั้น
- purchase history และ service history ช่วยให้ทีมหน้าร้านติดตามลูกค้าระยะยาวได้

### 7. Review Historical Context

- ผู้ใช้สามารถย้อนดู patient history, customer histories และ orders ที่เกี่ยวข้อง
- ข้อมูลทั้งหมดต้องอ่านกลับภายใต้ branch scope และสิทธิ์ที่ถูกต้อง

## Business Rules

### 1. Patient Profile

- ผู้ป่วยต้องมีข้อมูลระบุตัวตนอย่างน้อยที่เพียงพอต่อการค้นหาและติดตามประวัติ
- ข้อมูลสุขภาพถือเป็นข้อมูลอ่อนไหว ต้องใช้ภายใต้สิทธิ์ที่เหมาะสม
- หากระบบปิด feature ผู้ป่วยไว้ หน้าที่เกี่ยวข้องต้องไม่เปิดให้ใช้งาน

### 2. Allergy and Medical History

- ประวัติแพ้ยาเป็นข้อมูลสำคัญระดับ block/checkpoint ใน workflow การขาย
- severity ของ allergy มีผลต่อระดับการเตือน
- ข้อมูลโรคประจำตัวควรถูกใช้ร่วมกับการตัดสินใจของเภสัชกร แต่ไม่ควรถูกแก้ไขโดยไม่มีสิทธิ์

### 3. Patient-Linked Order History

- ทุกการขายที่ผูกกับ patient ต้องสามารถย้อนกลับมาดูประวัติได้
- ประวัตินี้ควรเชื่อมกับ order และ patient profile
- ประวัตินี้ใช้เพื่อการติดตามการรักษา

### 4. Customer CRM

- ลูกค้าใช้สำหรับติดตามประวัติการซื้อ การขายเครดิต หรือการตลาดพื้นฐาน
- ลูกค้ากับผู้ป่วยอาจเป็นคนเดียวกันได้ในเชิงธุรกิจ แต่ต้องแยก domain ตามวัตถุประสงค์การใช้งาน

## Validation Rules

- identifier สำคัญ เช่น เลขบัตรประชาชน ต้องมีรูปแบบสมเหตุสมผล
- ข้อมูลแพ้ยาต้องระบุสารหรือยาที่แพ้ได้ชัดเจน
- การเข้าถึงข้อมูลสุขภาพต้องถูกจำกัดตามสิทธิ์

## Edge Cases

- ผู้ป่วยไม่มีข้อมูลครบ แต่ต้องทำรายการขายด่วน ควรแยกกรณี guest patient ชัดเจน
- ข้อมูลผู้ป่วยซ้ำ ต้องมีวิธีจัดการ duplicate
- การลบข้อมูลไม่ควรทำให้ประวัติการขายเสีย integrity

## Expected Outcomes

- ใช้ข้อมูลผู้ป่วยช่วยลดความเสี่ยงทางคลินิก
- ค้นประวัติย้อนหลังได้ง่าย
- รองรับการติดตามการใช้ยาและบริการลูกค้าระยะยาว
