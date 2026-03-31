# 14. Dispensing Logs Contract

## เป้าหมาย

กำหนด contract สำหรับการบันทึกและอ่านข้อมูลการจ่ายยาที่ผูกกับผู้ป่วยและ order เพื่อใช้ในการติดตามประวัติการจ่ายยา การเตือน refill และรายงานเชิงคลินิก

## ฝั่งที่เกี่ยวข้อง

- Frontend หน้า patients, POS และ dashboard (refill reminders)
- Backend dispensing log services

## Contract Expectations

### 1. Create Dispensing Log

- ถูกสร้างโดยอัตโนมัติเมื่อ order ที่มีสินค้าผูก patient ถูกยืนยัน
- Request ต้องมี patient reference, product/order reference และข้อมูลการจ่ายที่จำเป็น
- Backend ต้อง validate ว่า patient และ order มีอยู่จริง

### 2. Dispensing Log List

- Frontend คาดหวัง list ที่ค้นหาและ filter ได้
- Response ควรมีข้อมูลสรุปของแต่ละรายการเพียงพอต่อการแสดงใน list

### 3. Dispensing Log Detail

- Frontend คาดหวังข้อมูลรายละเอียดเมื่อเปิดดู record เฉพาะ
- ต้องมีข้อมูล product, quantity, date, patient reference และ order reference

### 4. Dispensing Log by Patient

- Frontend คาดหวังประวัติการจ่ายยาของผู้ป่วยรายเดียว
- ใช้สำหรับหน้า patient detail และ refill reminder calculation
- Response ต้องเรียงตามวันที่จ่ายเพื่อให้ใช้ในการคำนวณ refill ได้

## Endpoints

| Method | Path | คำอธิบาย |
|---|---|---|
| POST | /dispensing-logs | สร้าง dispensing log |
| GET | /dispensing-logs | ดูรายการ dispensing logs |
| GET | /dispensing-logs/:id | ดูรายละเอียด dispensing log |
| GET | /dispensing-logs/patient/:patientId | ดูประวัติการจ่ายยาของผู้ป่วย |

## Error Cases

- patient not found
- invalid order reference
- insufficient permission to access medical data
- patient feature disabled
