BEGIN;

INSERT INTO auspicious_rule_sets (
    code, name_th, version, tradition, source_url, conflict_policy
) VALUES (
    'thai_wedding_tradition_v1',
    'ดิถีเรียงหมอนตามตำราโหราศาสตร์ไทย',
    '1.0.0',
    'โหราศาสตร์ไทย',
    'https://theluckyname.com/wedding-auspicious-riang-mon',
    '{"exclude_wins":true,"score_group_strategy":"highest_priority_only","notes":["ดิถีเรียงหมอนเป็นเกณฑ์มงคลเฉพาะสำหรับพิธีมงคลสมรสและส่งตัวเข้าหอ"]}'::jsonb
)
ON CONFLICT (code) DO UPDATE SET
    name_th = EXCLUDED.name_th,
    version = EXCLUDED.version,
    tradition = EXCLUDED.tradition,
    source_url = EXCLUDED.source_url,
    conflict_policy = EXCLUDED.conflict_policy,
    enabled = TRUE,
    updated_at = NOW();

WITH rule_set AS (
    SELECT id FROM auspicious_rule_sets WHERE code = 'thai_wedding_tradition_v1'
), seed(code, signal_group, name_th, description_th, meaning_th, score, priority, condition_json, source_note) AS (
    VALUES
    (
      'thai_wedding_riang_mon', 'wedding_riang_mon', 'วันเรียงหมอน (เคียงหมอน)',
      'ตรงกับวันดิถีเรียงหมอนมงคล (ขึ้น 7, 10, 13 ค่ำ และ แรม 4, 8, 10, 14 ค่ำ) เหมาะยิ่งสำหรับพิธีแต่งงาน ปูที่นอน และส่งตัวเจ้าสาวเข้าหอ',
      'เสริมความมั่นคงในชีวิตคู่ ครองเรือนยืนยาวจนแก่เฒ่า และมีความสุขความเจริญรุ่งเรือง',
      25, 500,
      '{"any":[{"all":[{"field":"lunar_phase","operator":"eq","value":"waxing"},{"field":"lunar_day","operator":"in","value":[7,10,13]}]},{"all":[{"field":"lunar_phase","operator":"eq","value":"waning"},{"field":"lunar_day","operator":"in","value":[4,8,10,14]}]}]}'::jsonb,
      'ตำราโหราศาสตร์ไทย: ดิถีเรียงหมอน (ขึ้น 7, 10, 13 ค่ำ และ แรม 4, 8, 10, 14 ค่ำ)'
    )
)
INSERT INTO auspicious_rules (
    rule_set_id, code, signal_group, name_th, description_th, meaning_th,
    effect, score, priority, condition_json, source_note
)
SELECT rule_set.id, seed.code, seed.signal_group, seed.name_th, seed.description_th, seed.meaning_th,
       'score', seed.score, seed.priority, seed.condition_json, seed.source_note
FROM rule_set CROSS JOIN seed
ON CONFLICT (code) DO UPDATE SET
    signal_group = EXCLUDED.signal_group,
    name_th = EXCLUDED.name_th,
    description_th = EXCLUDED.description_th,
    meaning_th = EXCLUDED.meaning_th,
    effect = EXCLUDED.effect,
    score = EXCLUDED.score,
    priority = EXCLUDED.priority,
    condition_json = EXCLUDED.condition_json,
    source_note = EXCLUDED.source_note,
    enabled = TRUE,
    updated_at = NOW();

-- Update isan_marriage_avoid_7 to target Friday 7 ค่ำ (which is strictly forbidden in folk texts) to allow auspicious Riang Mon on waxing 7
UPDATE auspicious_rules
SET condition_json = '{"all":[{"field":"weekday","operator":"eq","value":"friday"},{"field":"lunar_day","operator":"eq","value":7}]}'::jsonb,
    description_th = 'ตำราระบุให้เลี่ยงพิธีสมรสในวันศุกร์ที่ตรงกับ 7 ค่ำ',
    updated_at = NOW()
WHERE code = 'isan_marriage_avoid_7';

INSERT INTO auspicious_rule_activities (rule_id, activity_code)
SELECT id, 'wedding'
FROM auspicious_rules
WHERE code IN (
    'thai_wedding_riang_mon',
    'thai_kala_yok_athibodi',
    'thai_kala_yok_thongchai',
    'thai_kala_yok_ubat',
    'thai_kala_yok_lokawinat'
)
ON CONFLICT DO NOTHING;

COMMIT;
