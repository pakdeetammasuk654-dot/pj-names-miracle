BEGIN;

UPDATE auspicious_activities
SET name_th = 'เปิดกิจการหรือเปิดตัวสินค้า',
    icon = '🏪',
    description_th = 'เลือกวันเปิดร้าน เริ่มกิจการ เปิดตัวสินค้า เว็บไซต์ แอป เพจ หรือร้านออนไลน์',
    rule_scope_note_th = 'ใช้กฎทั่วไป พร้อมปุณณดิถีซึ่งตำราระบุว่าเหมาะกับการค้า',
    sort_order = 1,
    enabled = TRUE
WHERE code = 'business_opening';

UPDATE auspicious_activities
SET enabled = TRUE
WHERE code = 'career_start';

COMMIT;
