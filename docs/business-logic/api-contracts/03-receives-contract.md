# 03. Receives Contract

## เป้าหมาย

กำหนด contract สำหรับการสร้างและอ่านเอกสารรับสินค้า รวมถึงข้อมูลที่ frontend ใช้ใน receive workflow

## ฝั่งที่เกี่ยวข้อง

- Frontend หน้า receives
- Backend receive, lot, stock และ product history services

## Contract Expectations

### 1. Receive List and Detail

- Frontend คาดหวังรายการเอกสารรับสินค้าและข้อมูล detail ตามเอกสารที่เลือก
- Detail ควรมีรายการสินค้า supplier วันที่ และข้อมูลสรุปที่เกี่ยวข้อง

### 2. Create Receive

- Request ส่งเป็นเอกสารรับสินค้าทั้งชุด พร้อม item หลายรายการ
- Backend ต้อง validate ทั้งระดับเอกสารและระดับ item
- การตอบกลับควรมีข้อมูล receive ที่ถูกสร้างแล้ว หรือข้อมูลอย่างน้อยที่ทำให้ frontend refresh state ได้

### 3. Update and Partial Update

- `PUT /receives/:receiveId` — แก้ไขเอกสารรับสินค้าทั้งชุด
- `DELETE /receives/:receiveId` — cancel เอกสารรับสินค้าแบบเก็บเอกสารไว้ ไม่ใช่ hard delete
- `PATCH /receives/:receiveId/total-cost` — อัปเดตต้นทุนรวมของเอกสาร
- `PATCH /receives/:receiveId/items` — อัปเดตรายการสินค้าในเอกสาร
- `PATCH /receives/:receiveId/import` — นำเข้าข้อมูลรับสินค้าเข้าสู่ stock (trigger stock/lot/history update)
- เมื่อแก้ `items` backend ต้อง sync ทั้ง field `receive.items` และ collection `receive_items` ให้ตรงกัน เพราะ import flow และรายงาน ข.ย.9 อ่านจาก `receive_items`
- เมื่อแก้ `items` backend ควรคำนวณ `totalCost` ระดับเอกสารใหม่จากรายการล่าสุดอัตโนมัติ แทนการปล่อยให้ยอดรวมค้างจาก state เก่า

### 4. Data Integrity

- ถ้า item ใด item หนึ่งผิด ต้องไม่สร้าง receive ครึ่งทาง
- Backend ต้องรับประกันความสอดคล้องของ stock, lot และ history เมื่อ create หรือ import สำเร็จ
- Backend ต้องไม่ปล่อยให้ receive document กับ `receive_items` collection ถือข้อมูลคนละชุดหลังจาก update items
- Backend ต้องไม่ปล่อยให้ `totalCost` ของเอกสารไม่ตรงกับผลรวมจาก receive items ล่าสุด

## Endpoints

| Method | Path | คำอธิบาย |
|---|---|---|
| POST | /receives | สร้างเอกสารรับสินค้า |
| GET | /receives | ดูรายการเอกสารรับสินค้า |
| GET | /receives/:receiveId | ดูรายละเอียดเอกสาร |
| PUT | /receives/:receiveId | แก้ไขเอกสาร |
| DELETE | /receives/:receiveId | ยกเลิกเอกสารที่ยังไม่ถูก import |
| PATCH | /receives/:receiveId/total-cost | อัปเดตต้นทุนรวม |
| PATCH | /receives/:receiveId/items | อัปเดตรายการสินค้า |
| PATCH | /receives/:receiveId/import | นำเข้าสู่ stock |

## Error Cases

- invalid supplier
- invalid product
- quantity / price / expiry invalid
- branch context ไม่ถูกต้อง
- receive not found
- import ซ้ำ (ถ้ามีการป้องกัน)
- cancel imported receive not allowed
- receive items out of sync with document state ต้องไม่เกิดขึ้นใน flow ปกติ
