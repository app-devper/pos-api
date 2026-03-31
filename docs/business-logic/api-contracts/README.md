# API Contract Documentation

เอกสารชุดนี้อธิบายข้อตกลงการสื่อสารระหว่าง `pos-web` และ `pos-api` ในระดับโดเมน โดยเน้นความคาดหวังเชิงธุรกิจและพฤติกรรมของ request/response มากกว่าการอธิบายโค้ดภายใน

## เอกสารในชุดนี้

- `01-auth-and-context-contract.md`
- `02-products-contract.md`
- `03-receives-contract.md`
- `04-pos-orders-contract.md`
- `05-reports-contract.md`
- `06-patients-and-customers-contract.md`
- `07-settings-contract.md`
- `08-promotions-contract.md`
- `09-stock-transfers-contract.md`
- `10-branches-contract.md`
- `11-employees-contract.md`
- `12-suppliers-contract.md`
- `13-categories-contract.md`
- `14-dispensing-logs-contract.md`
- `15-customer-histories-contract.md`
- `16-dashboard-contract.md`

## เนื้อหาที่แต่ละไฟล์ควรครอบคลุม

- เป้าหมายของ contract
- ฝั่งที่เรียกใช้และฝั่งที่ให้บริการ
- ข้อมูลนำเข้าหลัก
- ข้อมูลตอบกลับที่ frontend คาดหวัง
- validation และ authorization ที่ backend ต้องบังคับ
- error cases ที่ frontend ต้องรองรับ
- หมายเหตุเรื่องความสอดคล้องกับ business logic docs
