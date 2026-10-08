BEGIN;

UPDATE auspicious_activities
SET name_th = 'เปลี่ยนชื่อหรือตั้งชื่อสกุล',
    description_th = 'เลือกวันเริ่มใช้ชื่อใหม่ ยื่นเอกสาร หรือประกาศใช้ชื่อสกุล',
    safety_note_th = ''
WHERE code = 'name_change';

UPDATE auspicious_activities
SET name_th = 'เปิดร้านหรือกิจการ',
    description_th = 'หาวันเหมาะสำหรับเปิดร้าน เริ่มกิจการ หรือเริ่มค้าขาย'
WHERE code = 'business_opening';

UPDATE auspicious_activities
SET enabled = TRUE
WHERE code IN ('important_documents', 'product_launch');

COMMIT;
