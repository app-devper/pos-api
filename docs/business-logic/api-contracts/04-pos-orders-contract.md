# 04. POS Orders Contract

## เป้าหมาย

กำหนด contract สำหรับการสร้าง order, การชำระเงิน, และข้อมูลที่เกี่ยวข้องกับการขายหน้าร้าน

## ฝั่งที่เกี่ยวข้อง

- Frontend หน้า POS
- Backend order, payment, stock deduction และ document-related services

## Contract Expectations

### 1. Product Lookup for POS

- Frontend คาดหวังข้อมูลที่ค้นหาเร็วพอสำหรับงานหน้าร้าน
- ผลลัพธ์ควรมีข้อมูลสินค้า หน่วย ราคา และข้อมูลเพียงพอสำหรับเพิ่มเข้าตะกร้า

### 2. Submit Order

- Request ต้องรวม cart items, payment data, customer/patient references และ compliance data เมื่อจำเป็น
- Compliance fields ในกรณียาควบคุม: `pharmacistName`, `licenseNo`, `prescriberName`, `buyerName`, `buyerIdCard`
- Backend ต้อง validate branch authorization, payment total และความสามารถในการตัด stock จาก `item.Stocks`
- ปัจจุบัน compliance fields ถูกเก็บตาม payload ที่ client ส่งมา และ frontend เป็นตัว enforce rule ว่าต้องกรอกเมื่อใด
- Response ควรมีข้อมูล order ที่บันทึกสำเร็จและ identifiers ที่ใช้กับเอกสารหลังการขาย
- Order entity เก็บ compliance data โดยตรง เพื่อรองรับรายงาน ข.ย. 10–13 ที่ดึงจาก orders

### 2.1 Record Sale (`POST /orders` ที่มี `saleId`)

Till ส่งเฉพาะสิ่งที่แคชเชียร์เลือก server เป็นผู้คิดราคาและตัด Stock เองจาก Stock ปัจจุบันใน transaction เดียว (order, items, payments, stock, sold first, product history — สำเร็จทั้งหมดหรือไม่บันทึกเลย)

```json
{
  "saleId": "uuid สร้างโดย till ครั้งเดียวต่อ Sale",
  "type": "CASH",
  "payments": [{ "amount": 100, "type": "CASH" }],
  "items": [
    {
      "productId": "...",
      "unitId": "...",
      "quantity": 2,
      "priceType": "General | Regular | Wholesaler | Stock",
      "stockId": "ล็อตที่แคชเชียร์เลือก (optional)",
      "discount": 1,
      "allowOversell": false
    }
  ],
  "customerCode": "...", "customerName": "...", "patientId": "...",
  "pharmacistName": "...", "licenseNo": "...", "prescriberName": "...", "buyerName": "...", "buyerIdCard": "..."
}
```

- ราคา: `priceType: "Stock"` ใช้ราคาของ Stock (ล็อตที่เลือก หรือล็อตแรกตาม sequence ที่ยังมีของ) ถ้ามี; ไม่เช่นนั้นใช้ price list ของ customer type นั้น, ถ้าไม่มีใช้ price list แรกของ Unit, ถ้าไม่มีเลยราคา 0
- `discount` ต่อหน่วย ถูกจำกัดให้อยู่ระหว่าง 0 ถึงราคาต่อหน่วย; ไม่มีส่วนลดระดับบิล
- ตัด Stock: ล็อตที่เลือกก่อน แล้วล็อตอื่นใน branch ตาม sequence; ส่วนที่ขาดเข้า Sold first ยกเว้น `allowOversell` และบรรทัดได้ตัดล็อตจริงแล้ว — ส่วนที่ขาดผูกกับล็อตสุดท้าย (`oversoldQty`)
- ต้นทุนคิดจากล็อตที่ตัดจริง (ล็อตไม่มีต้นทุนหรือ Sold first ใช้ต้นทุนของ Unit)
- server คำนวณ `total`, `totalCost`, `discount`, `change` เอง; ยอดชำระรวมน้อยกว่า total → `400 OR-400-001`
- ส่ง `saleId` เดิมซ้ำด้วยเนื้อหาเดิม → คืน Order เดิม (ไม่ตัด stock ซ้ำ ไม่ใช้เลข order ใหม่); เนื้อหาต่างกัน → `409 OR-409-001`
- Response: `{ "data": Order, "stocks": [ProductStock ที่ถูกตัด] }`
- Request ที่ไม่มี `saleId` ยังใช้ payload เดิม (till ตัดสินราคาและล็อตเอง) จนกว่า till รุ่นเก่าจะหมด

### 2.2 ความหมายของเงินใน Order

- `item.price` = ยอดบรรทัด (ราคาต่อหน่วย × จำนวน) ก่อนหักส่วนลด
- `item.costPrice` = ต้นทุนของทั้งบรรทัด
- `item.discount` = ส่วนลดต่อหน่วย; ลูกค้าจ่าย `price − discount × quantity`
- `order.total` = Σ ยอดที่ลูกค้าจ่ายของบรรทัดที่ยังไม่ยกเลิก, `order.totalCost` = Σ `costPrice`, `order.discount` = Σ `discount × quantity` — ยกเลิกบรรทัดแล้วคำนวณใหม่ตามนี้
- เงินคืนของ Return ต่อหน่วยไม่เกินที่ลูกค้าจ่ายต่อหน่วย (`price / quantity − discount`)

### 3. Post-Order Data

- Frontend คาดหวังข้อมูลอ้างอิงสำหรับ receipt, tax invoice, labels หรือ report-related follow-up actions
- Backend ต้องสะท้อนสถานะคำสั่งขายที่ชัดเจน
- Order ใหม่ถูกคาดหวังให้อยู่ในสถานะ `CONFIRMED`
- การยกเลิกหลังบันทึกต้องเก็บเอกสารไว้พร้อมสถานะ `CANCELLED` และข้อมูล audit ที่เพียงพอ เช่น `cancelReason`

### 4. Order Item Management

- Frontend สามารถดึง order items แยกจาก order หลักได้
- รองรับการค้นหา items ตาม product เพื่อใช้ในรายงานหรือ history
- รองรับการยกเลิก item หลังบันทึกโดยไม่ทำลายประวัติย้อนหลัง
- order items ที่ถูกยกเลิกต้องไม่ถูกนับใน total/report ปกติ

### 5. Customer Code Update

- Frontend สามารถอัปเดต customer code ของ order ที่สร้างแล้วได้ผ่าน PATCH
- ใช้ในกรณีที่ผูกลูกค้าภายหลังการสร้าง order

### 6. Cancel Actions

- `DELETE /orders/:orderId`, `DELETE /orders/items/:itemId`, และ `DELETE /orders/:orderId/products/:productId` ใช้เป็น cancel semantics ไม่ใช่ hard delete
- Request body สามารถส่ง `{"reason":"..."}` เพื่อเก็บเหตุผลการยกเลิกสำหรับ audit ได้
- ถ้าไม่ส่ง `reason` ระบบยังต้องทำงานได้เพื่อคง backward compatibility กับ client เดิม
- ถ้าส่ง JSON body มาแต่ parse ไม่ได้ ระบบต้องตอบ `400 bad request` แทนการ ignore body แล้ว cancel ต่อ
- การ cancel ต้อง reverse stock อย่างสอดคล้อง และ mark payment/item/order เป็น `CANCELLED`

## Endpoints

| Method | Path | คำอธิบาย |
|---|---|---|
| POST | /orders | สร้าง order ใหม่ |
| GET | /orders | ดูรายการ orders |
| GET | /orders/:orderId | ดูรายละเอียด order |
| DELETE | /orders/:orderId | ยกเลิก order ทั้งใบ พร้อมรับ optional body `{ "reason": "..." }` |
| PATCH | /orders/:orderId/customer-code | อัปเดต customer code |
| GET | /orders/items | ดูรายการ order items (range) |
| GET | /orders/items/:itemId | ดูรายละเอียด order item |
| DELETE | /orders/items/:itemId | ยกเลิก order item พร้อมรับ optional body `{ "reason": "..." }` |
| DELETE | /orders/:orderId/products/:productId | ยกเลิก order item ตาม orderId+productId พร้อมรับ optional body `{ "reason": "..." }` |
| GET | /orders/items/products/:productId | ค้นหา items ตาม product |
| GET | /orders/item-details/products/:productId | ค้นหา item details ตาม product |
| GET | /orders/customers/:customerCode | ดู orders ของลูกค้า |

## Cancel Payload

```json
{
  "reason": "customer changed mind"
}
```

- `reason` เป็น optional string
- ค่านี้จะถูกบันทึกใน order/item/payment ที่ถูก cancel เพื่อใช้ audit ย้อนหลัง

## Error Cases

- stock insufficient
- compliance data missing for controlled drugs
- invalid payment total
- `saleId` ซ้ำกับ Sale อื่น (`409 OR-409-001`)
- malformed cancel action payload
- invalid product or unit reference
- branch context ไม่ถูกต้อง
- order not found
- invalid order state for requested action
- session expired during submit
