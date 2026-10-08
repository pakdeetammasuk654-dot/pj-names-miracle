BEGIN;

UPDATE auspicious_rules
SET description_th = 'วันพฤหัสบดีที่ตรง 5, 10 หรือ 15 ค่ำ ตำราระบุว่าเป็นดิถีแห่งความบริบูรณ์ เหมาะแก่การตั้งชื่อ เริ่มต้นสิ่งมงคล และการค้าขาย',
    meaning_th = 'ปุณณดิถีหมายถึงดิถีแห่งความเต็มเปี่ยมและบริบูรณ์ สื่อถึงความเจริญงอกงาม ความสำเร็จ และความสมบูรณ์พูนสุข'
WHERE code = 'isan_punnaditthi_business';

UPDATE auspicious_activities
SET rule_scope_note_th = 'ใช้กฎทั่วไป พร้อมปุณณดิถีซึ่งตำราระบุว่าเป็นดิถีแห่งความบริบูรณ์ เหมาะแก่การตั้งชื่อและเริ่มชีวิตใหม่'
WHERE code = 'name_change';

INSERT INTO auspicious_rule_activities (rule_id, activity_code)
SELECT id, 'name_change'
FROM auspicious_rules
WHERE code = 'isan_punnaditthi_business'
ON CONFLICT DO NOTHING;

COMMIT;
