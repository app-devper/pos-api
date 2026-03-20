# 08. Report Document Lifecycle

## วัตถุประสงค์

อธิบายวงจรการสร้างและใช้งานเอกสารรายงานและเอกสาร export เพื่อให้เข้าใจความสัมพันธ์ระหว่างข้อมูลดิบ การ preview การพิมพ์ และไฟล์ที่ผู้ใช้ดาวน์โหลด

## สถานะหลักเชิงพฤติกรรม

- `INPUT_PENDING` — ผู้ใช้ยังกรอกเงื่อนไขไม่ครบ จึงยังสร้างเอกสารไม่ได้
- `DATA_READY` — Backend คืนข้อมูลดิบที่พร้อมใช้สร้างเอกสารแล้ว
- `PREVIEW_READY` — Frontend render เอกสารหรือ preview ได้แล้ว
- `EXPORTED` — ผู้ใช้สั่งพิมพ์หรือดาวน์โหลดไฟล์แล้ว
- `FAILED` — การโหลดข้อมูลหรือการสร้างเอกสารล้มเหลว

## Transition หลัก

### 1. `INPUT_PENDING -> DATA_READY`

#### Trigger

- ผู้ใช้กรอก params ครบและ Frontend เรียก data endpoint สำเร็จ

#### Preconditions

- input ถูกต้อง
- ผู้ใช้มีสิทธิ์เข้าถึงรายงานนั้น
- branch context ถูกต้อง

#### Backend Effects

- query ข้อมูลและคืน payload ที่พร้อมสำหรับ render
- สำหรับ CSV flow อาจเตรียมไฟล์แทน document model

#### Frontend Behavior

- เก็บข้อมูลใน state สำหรับ preview และ export
- เปิดปุ่ม preview / print / export ที่เกี่ยวข้อง

### 2. `DATA_READY -> PREVIEW_READY`

#### Trigger

- Frontend map data เป็น document model และ render สำเร็จ

#### Effects

- ผู้ใช้มองเห็นเอกสารบนจอ
- สามารถตรวจข้อมูลก่อนพิมพ์ได้

### 3. `PREVIEW_READY -> EXPORTED`

#### Trigger

- ผู้ใช้กด print, save as PDF, หรือ download CSV

#### Effects

- browser print flow หรือ file download ถูกเรียกใช้งาน
- เอกสารที่ผู้ใช้ได้มาต้องอิงจากข้อมูลล่าสุดที่โหลดสำเร็จ

### 4. `INPUT_PENDING/DATA_READY/PREVIEW_READY -> FAILED`

#### Trigger

- input ไม่ครบ
- ไม่มีสิทธิ์
- server response fail
- browser block print window หรือ render fail

#### Frontend Behavior

- แสดง error state ชัดเจน
- อนุญาตให้ผู้ใช้แก้ input หรือ retry ได้

## ข้อควรระวัง

- ห้ามใช้ state เก่ามาสร้าง PDF ถ้าผู้ใช้เปลี่ยน filter แล้วแต่ยังไม่ได้ reload data
- รายงาน frontend-generated PDF ต้องผูกกับ data endpoint เดียวกับ preview เพื่อหลีกเลี่ยงข้อมูลไม่ตรงกัน
- CSV และ PDF ของรายงานเดียวกันอาจมาจากคนละ flow แต่ต้องสะท้อนชุดข้อมูลธุรกิจเดียวกัน
