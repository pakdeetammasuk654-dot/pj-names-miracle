BEGIN;

INSERT INTO auspicious_activities (
    code, name_th, icon, description_th, rule_scope_note_th, safety_note_th,
    sort_order, enabled, category_code
) VALUES (
    'negotiation',
    'เจรจาต่อรอง ทำสัญญา หรือปิดการขาย',
    '🤝',
    'เลือกวันสำหรับเจรจาธุรกิจ ขอความร่วมมือ เสนอเงื่อนไข เซ็นสัญญา หรือปิดการขาย',
    'ใช้กฎวันมงคลตามตำราอีสานและปุณณดิถีสำหรับการค้า ร่วมกับกาลโยคเพื่อประกอบการตัดสินใจ',
    'ควรเตรียมข้อมูล เงื่อนไข และตรวจสอบข้อกฎหมายของสัญญาให้ครบถ้วนก่อนเสมอ',
    2, TRUE, 'work_business'
)
ON CONFLICT (code) DO UPDATE SET
    name_th = EXCLUDED.name_th,
    icon = EXCLUDED.icon,
    description_th = EXCLUDED.description_th,
    rule_scope_note_th = EXCLUDED.rule_scope_note_th,
    safety_note_th = EXCLUDED.safety_note_th,
    sort_order = EXCLUDED.sort_order,
    enabled = TRUE,
    category_code = EXCLUDED.category_code;

INSERT INTO auspicious_rule_sets (
    code, name_th, version, tradition, source_url, conflict_policy
) VALUES (
    'thai_kala_yok_v1',
    'กาลโยคตามคัมภีร์โหราศาสตร์ไทย',
    '1.0.0',
    'โหราศาสตร์ไทย',
    'https://nsquare.co/tools/kala-yok-calculator',
    '{"exclude_wins":true,"score_group_strategy":"highest_priority_only","notes":["กาลโยคเป็นสัญญาณประกอบการเลือกวัน","แสดงวันดีและวันควรระวังพร้อมกันเมื่อทับซ้อน"]}'::jsonb
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
    SELECT id FROM auspicious_rule_sets WHERE code = 'thai_kala_yok_v1'
), seed(code, signal_group, name_th, description_th, meaning_th, score, priority, condition_json, source_note) AS (
    VALUES
    (
      'thai_kala_yok_athibodi', 'kala_yok_athibodi', 'วันอธิบดี',
      'ตรงวันอธิบดีตามกาลโยค ซึ่งใช้ประกอบการเลือกวันสำหรับเรื่องบุคคล อำนาจ และสัญญาสำคัญ',
      'วันอธิบดีเป็นวันมงคลด้านอำนาจ บารมี และการดำเนินเรื่องสำคัญกับบุคคล',
      12, 420, '{"field":"is_kala_yok_athibodi","operator":"eq","value":true}'::jsonb,
      'คัมภีร์กาลโยค: เกณฑ์อธิบดี = จุลศักราช mod 498 แล้วหารฐานวัน 7'
    ),
    (
      'thai_kala_yok_thongchai', 'kala_yok_thongchai', 'วันธงชัย',
      'ตรงวันธงชัยตามกาลโยค ซึ่งใช้ประกอบการเลือกวันเริ่มต้นงานที่ต้องการความสำเร็จและชัยชนะ',
      'วันธงชัยเป็นวันมงคลที่สื่อถึงความสำเร็จและชัยชนะของการดำเนินงาน',
      8, 400, '{"field":"is_kala_yok_thongchai","operator":"eq","value":true}'::jsonb,
      'คัมภีร์กาลโยค: เกณฑ์ธงชัย = (จุลศักราช × 10) + 3 แล้วหารฐานวัน 7'
    ),
    (
      'thai_kala_yok_ubat', 'kala_yok_ubat', 'วันอุบาทว์',
      'ตรงวันอุบาทว์ตามกาลโยค จึงควรวางแผนและตรวจทานเงื่อนไขการเจรจาเป็นพิเศษ',
      'วันอุบาทว์เป็นสัญญาณเตือนเรื่องอุปสรรคสำหรับกิจกรรมสำคัญที่เกี่ยวข้องกับบุคคล',
      -10, 410, '{"field":"is_kala_yok_ubat","operator":"eq","value":true}'::jsonb,
      'คัมภีร์กาลโยค: เกณฑ์อุบาทว์ = (จุลศักราช × 10) + 2 แล้วหารฐานวัน 7'
    ),
    (
      'thai_kala_yok_lokawinat', 'kala_yok_lokawinat', 'วันโลกาวินาศ',
      'ตรงวันโลกาวินาศตามกาลโยค จึงควรระวังความขัดข้องและไม่เร่งตัดสินใจในเงื่อนไขสำคัญ',
      'วันโลกาวินาศเป็นสัญญาณเตือนเรื่องความวุ่นวายและความเสียหาย จึงใช้ลดความเหมาะสมของวัน',
      -8, 410, '{"field":"is_kala_yok_lokawinat","operator":"eq","value":true}'::jsonb,
      'คัมภีร์กาลโยค: เกณฑ์โลกาวินาศ = จุลศักราช + 1120 แล้วหารฐานวัน 7'
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

INSERT INTO auspicious_rule_activities (rule_id, activity_code)
SELECT id, 'negotiation'
FROM auspicious_rules
WHERE code IN (
    'isan_punnaditthi_business',
    'thai_kala_yok_athibodi',
    'thai_kala_yok_thongchai',
    'thai_kala_yok_ubat',
    'thai_kala_yok_lokawinat'
)
ON CONFLICT DO NOTHING;

COMMIT;
