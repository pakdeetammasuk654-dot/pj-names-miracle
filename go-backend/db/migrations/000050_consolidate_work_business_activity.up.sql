BEGIN;

UPDATE auspicious_activities
SET name_th = 'เริ่มงานใหม่ รับตำแหน่ง หรือเปิดกิจการ',
    icon = '💼',
    description_th = 'เลือกวันเริ่มงานวันแรก รับตำแหน่งใหม่ เปิดร้าน เริ่มธุรกิจ หรือเปิดตัวสินค้า เว็บไซต์ แอป เพจ และร้านออนไลน์',
    rule_scope_note_th = 'ใช้กฎทั่วไป พร้อมปุณณดิถีซึ่งตำราระบุว่าเหมาะกับการค้า ธุรกิจ และความก้าวหน้า',
    sort_order = 1,
    enabled = TRUE
WHERE code = 'business_opening';

UPDATE auspicious_activities
SET enabled = FALSE
WHERE code IN ('career_start', 'product_launch');

COMMIT;
