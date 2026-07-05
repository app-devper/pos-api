# 09. Stock Transfers Contract

## เป้าหมาย

กำหนด contract สำหรับการสร้าง อ่าน และเปลี่ยนสถานะคำขอโอนสต็อกระหว่างสาขา โดยให้ frontend และ backend เห็นสถานะเดียวกัน

## ฝั่งที่เกี่ยวข้อง

- Frontend หน้า stock transfers
- Backend stock transfer services

## Contract Expectations

### 1. Transfer List / Detail

- Frontend คาดหวังรายการคำขอโอนพร้อมสถานะและข้อมูลสรุป
- Detail ต้องมีต้นทาง ปลายทาง รายการสินค้า และข้อมูลการอนุมัติ/ปฏิเสธเมื่อมี

### 2. Create Transfer Request

- Request ต้องมี branch ต้นทาง ปลายทาง และ items ที่ถูกต้อง
- Backend ต้อง validate branch pair, quantity, product references และเงื่อนไขพื้นฐานอื่น
- Response ควรมี transfer record ล่าสุดในสถานะเริ่มต้นที่ชัดเจน

### 3. Approve / Reject / Cancel

- Frontend คาดหวัง endpoint หรือ action contract ที่เปลี่ยนสถานะได้ชัดเจน
- Backend ต้อง validate stock ณ เวลาที่อนุมัติ ไม่ใช่เฉพาะตอนสร้างคำขอ
- Response ควรสะท้อนสถานะล่าสุดและผลกระทบเชิงธุรกิจ

## Error Cases

- same source and destination branch
- insufficient stock at approval time
- invalid item references
- invalid state transition เช่น approve ซ้ำ
