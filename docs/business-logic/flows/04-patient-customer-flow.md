# 04. Patient and Customer Flow

## เป้าหมาย

บริหารข้อมูลผู้ป่วยและลูกค้าให้ใช้งานได้จริงในงานขาย การติดตามประวัติ และการดูแลต่อเนื่อง โดยไม่ทำให้ข้อมูลสุขภาพและข้อมูลเชิงพาณิชย์ปะปนกันผิด domain

## Actors

- พนักงานขาย
- เภสัชกร
- เจ้าหน้าที่หน้าร้าน
- Frontend หน้า patients / customers / POS
- Backend patient / customer / history services

## Preconditions

- feature ผู้ป่วยเปิดใช้งาน หากจะใช้ patient flow
- ผู้ใช้มีสิทธิ์เข้าถึงข้อมูลตาม role

## Main Flow A: Customer Management

### Step 1: เปิดหน้าลูกค้า

- Frontend โหลด customer list พร้อม search/filter
- ผู้ใช้เลือกดูข้อมูลหรือสร้างลูกค้าใหม่

### Step 2: Create / Update Customer

- ผู้ใช้กรอกข้อมูลพื้นฐาน เช่น ชื่อ เบอร์โทร ประเภทลูกค้า
- Frontend ตรวจข้อมูลบังคับก่อนส่ง
- Backend ตรวจรูปแบบข้อมูลและบันทึก

### Step 3: Purchase History

- ผู้ใช้เลือกดูประวัติการซื้อ
- Frontend โหลด order history ของลูกค้า
- สามารถดูบน panel/page และพิมพ์รายงานประวัติลูกค้าได้จาก frontend

## Main Flow B: Patient Management

### Step 1: เปิดหน้าผู้ป่วย

- Frontend โหลดรายการผู้ป่วยและข้อมูลย่อ
- ผู้ใช้ค้นหาและเลือก record ที่ต้องการ

### Step 2: Update Medical Data

- ผู้ใช้แก้ไข allergy, chronic diseases, notes และข้อมูลสุขภาพที่เกี่ยวข้อง
- Frontend ต้องแยกส่วนข้อมูลทั่วไปกับข้อมูลสุขภาพให้ชัดเจน
- Backend บันทึกข้อมูลภายใต้สิทธิ์ที่เหมาะสม

### Step 3: Use Patient in POS

- ผู้ใช้ผูก patient กับการขาย
- Frontend ใช้ข้อมูลนี้เพื่อเช็ค allergy / interaction / history
- Backend ใช้ patient reference ในการบันทึก order และ dispensing log ที่เกี่ยวข้อง

## Decision Points

- ลูกค้าเป็น guest หรือมี profile ในระบบ
- ผู้ใช้ต้องการแค่ customer record หรือ patient record ที่มีข้อมูลสุขภาพ
- feature ผู้ป่วยเปิดหรือปิด

## Error Flow

- duplicate data → ต้องให้ผู้ใช้ตรวจสอบก่อนสร้าง record ใหม่
- ไม่มีสิทธิ์ดูข้อมูลสุขภาพ → แสดง restricted state
- ข้อมูล patient ไม่ครบ → อาจทำรายการขายได้ในบางกรณี แต่ต้องไม่หลอกว่ามีการตรวจสอบครบแล้ว

## Expected Outcome

- ค้นหาและติดตามลูกค้า/ผู้ป่วยได้ง่าย
- ใช้ข้อมูลผู้ป่วยช่วยเพิ่มความปลอดภัยของการจ่ายยา
- ประวัติย้อนหลังพร้อมใช้ทั้งในงานบริการและรายงาน
