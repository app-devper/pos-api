# 09. Stock Transfer Flow

## เป้าหมาย

โอนสต็อกระหว่างสาขาได้อย่างปลอดภัย โดยสถานะคำขอและยอดคงเหลือต้องสอดคล้องกันตลอด flow

## Actors

- ADMIN
- SUPER
- Frontend หน้า stock transfers
- Backend stock transfer services

## Preconditions

- ผู้ใช้มีสิทธิ์จัดการคำขอโอน
- มี branch context ที่เกี่ยวข้อง

## Main Flow

1. ผู้ใช้เปิดหน้าคำขอโอนสต็อก
2. Frontend โหลด transfer list, branch data และข้อมูลสินค้าที่จำเป็น
3. ผู้ใช้สร้างคำขอโอนโดยระบุต้นทาง ปลายทาง และรายการสินค้า
4. Frontend ตรวจ validation เบื้องต้น เช่น quantity และ branch selection
5. Backend สร้างคำขอในสถานะ pending
6. ผู้มีสิทธิ์ตรวจคำขอและเลือก approve / reject / cancel
7. หาก approve Backend ย้าย stock และอัปเดตสถานะ
8. Frontend refresh ข้อมูลล่าสุดให้ผู้ใช้เห็น

## Error Flow

- stock ไม่พอในตอนอนุมัติ → reject action พร้อมเหตุผล
- approve/reject ซ้ำ → ต้องถูกป้องกันทั้ง frontend และ backend
- ข้อมูล branch หรือสินค้าไม่ถูกต้อง → ไม่ให้สร้างคำขอ

## Expected Outcome

- คำขอโอนมีสถานะชัดเจน
- stock ต้นทางและปลายทางถูกต้องหลังอนุมัติ
- ผู้ใช้เข้าใจเหตุผลเมื่อคำขอถูกปฏิเสธหรือยกเลิก
