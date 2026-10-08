BEGIN;

UPDATE auspicious_activities
SET rule_scope_note_th = 'ใช้กฎทั่วไป พร้อมปุณณดิถีซึ่งตำราระบุว่าเหมาะกับการค้าและธุรกิจ'
WHERE code = 'brand_naming';

INSERT INTO auspicious_rule_activities (rule_id, activity_code)
SELECT id, 'brand_naming'
FROM auspicious_rules
WHERE code = 'isan_punnaditthi_business'
ON CONFLICT DO NOTHING;

COMMIT;
