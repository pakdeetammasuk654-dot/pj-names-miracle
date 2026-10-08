BEGIN;

UPDATE auspicious_rules
SET description_th = 'วันพฤหัสบดีที่ตรง 5, 10 หรือ 15 ค่ำ ตำราระบุว่าเหมาะกับการค้าและการเพาะปลูก',
    meaning_th = 'วันพฤหัสบดีที่ตรง 5, 10 หรือ 15 ค่ำ ตำราระบุว่าเหมาะกับการค้าและการเพาะปลูก'
WHERE code = 'isan_punnaditthi_business';

UPDATE auspicious_activities
SET rule_scope_note_th = ''
WHERE code = 'name_change';

DELETE FROM auspicious_rule_activities
WHERE activity_code = 'name_change'
  AND rule_id IN (SELECT id FROM auspicious_rules WHERE code = 'isan_punnaditthi_business');

COMMIT;
