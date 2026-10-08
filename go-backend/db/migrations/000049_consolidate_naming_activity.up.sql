BEGIN;

UPDATE auspicious_activities
SET name_th = 'ตั้งชื่อ เปลี่ยนชื่อ นามสกุล หรือชื่อธุรกิจ',
    description_th = 'เลือกวันเริ่มต้นใช้ชื่อใหม่ ตั้งชื่อ-เปลี่ยนชื่อตัว นามสกุล ชื่อร้าน แบรนด์ บริษัท หรือยื่นเอกสารราชการ',
    rule_scope_note_th = 'ใช้กฎทั่วไป พร้อมปุณณดิถีซึ่งตำราระบุว่าเป็นดิถีแห่งความบริบูรณ์ เหมาะแก่การตั้งชื่อและเริ่มต้นสิ่งมงคล',
    safety_note_th = 'ควรตรวจวันทำการ เวลารับเรื่องของหน่วยงานราชการ และเตรียมเอกสารให้พร้อมก่อนเลือกวัน',
    sort_order = 1,
    enabled = TRUE
WHERE code = 'name_change';

UPDATE auspicious_activities
SET enabled = FALSE
WHERE code IN ('brand_naming', 'important_documents');

INSERT INTO auspicious_rule_activities (rule_id, activity_code)
SELECT id, 'name_change'
FROM auspicious_rules
WHERE code = 'isan_punnaditthi_business'
ON CONFLICT DO NOTHING;

COMMIT;
