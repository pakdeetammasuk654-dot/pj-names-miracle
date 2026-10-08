BEGIN;

UPDATE auspicious_activities
SET name_th = 'เปลี่ยนชื่อ ตั้งชื่อสกุล หรือยื่นเอกสารสำคัญ',
    description_th = 'เลือกวันเริ่มใช้ชื่อใหม่ เปลี่ยนชื่อในเอกสารราชการ หรือดำเนินเรื่องสำคัญ',
    safety_note_th = 'ควรตรวจวันทำการ เวลารับเรื่อง และข้อกำหนดของหน่วยงานก่อนเลือกวัน'
WHERE code = 'name_change';

UPDATE auspicious_activities
SET name_th = 'เปิดกิจการหรือเปิดตัวสินค้า',
    description_th = 'เลือกวันเปิดร้าน เริ่มกิจการ เปิดตัวสินค้า เว็บไซต์ แอป เพจ หรือร้านออนไลน์'
WHERE code = 'business_opening';

UPDATE auspicious_activities
SET enabled = FALSE
WHERE code IN ('important_documents', 'product_launch');

COMMIT;
