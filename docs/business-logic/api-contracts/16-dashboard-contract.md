# 16. Dashboard Contract

## เป้าหมาย

กำหนด contract สำหรับข้อมูลสรุป กราฟ และ alerts ที่หน้า Dashboard ใช้แสดงภาพรวมของร้านในระดับสาขา

## ฝั่งที่เกี่ยวข้อง

- Frontend หน้า dashboard
- Backend dashboard, order, stock และ product services

## Contract Expectations

### 1. Summary

- Frontend คาดหวังข้อมูลสรุปยอดขาย จำนวน order ต้นทุน และกำไรตามช่วงเวลา
- Response ต้องผูกกับ branch context ของผู้ใช้

### 2. Daily / Monthly Chart

- Frontend คาดหวังข้อมูล time-series สำหรับ render กราฟยอดขาย
- Response ต้องมี date, totalOrders, totalRevenue, totalCost, totalProfit ต่อช่วง

### 3. Low Stock

- Frontend คาดหวังรายการสินค้าที่ stock ต่ำกว่า min stock
- ใช้เป็น alert section บน dashboard

### 4. Stock Report

- Frontend คาดหวังข้อมูลสรุป stock ทั้งหมดสำหรับภาพรวม
- Response ควรมี product name, quantity, cost summary

### 5. Expiring Products

- Frontend คาดหวังรายการสินค้าที่ใกล้หมดอายุตามจำนวนวันที่กำหนด
- ใช้เป็น alert section บน dashboard

### 6. Refill Reminders

- Frontend คาดหวังรายการผู้ป่วยที่ใกล้ถึงกำหนดเติมยา
- ปกติ frontend ควรแสดงเฉพาะเมื่อ patient feature ถูกเปิดตาม config ที่อ่านจาก settings โดย backend ไม่ได้ gate endpoint นี้จาก toggle ดังกล่าวโดยตรง

### 7. ABC Analysis

- Frontend คาดหวังรายการสินค้าจัดกลุ่ม A/B/C ตามยอดขายสะสม
- ใช้เพื่อวางแผนสต็อกและการสั่งซื้อ

### 8. Dead Stock

- Frontend คาดหวังรายการสินค้าที่ไม่มีการเคลื่อนไหวเกินช่วงเวลาที่กำหนด
- ใช้เป็นข้อมูลวิเคราะห์สินค้าค้างสต็อก

## Endpoints

| Method | Path | คำอธิบาย |
|---|---|---|
| GET | /dashboard/summary | สรุปยอดขาย ต้นทุน กำไร |
| GET | /dashboard/daily-chart | กราฟยอดขายรายวัน |
| GET | /dashboard/monthly-chart | กราฟยอดขายรายเดือน |
| GET | /dashboard/low-stock | รายการสินค้า stock ต่ำ |
| GET | /dashboard/stock-report | รายงานสรุป stock |
| GET | /dashboard/expiring | สินค้าใกล้หมดอายุ |
| GET | /dashboard/abc-analysis | วิเคราะห์ ABC |
| GET | /dashboard/dead-stock | สินค้าค้างสต็อก |

## Error Cases

- invalid date range
- unauthorized access
- branch context ไม่ถูกต้อง
- refill reminders not available in current frontend feature configuration
