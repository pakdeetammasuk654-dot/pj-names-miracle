BEGIN;

CREATE TABLE IF NOT EXISTS auspicious_rule_sets (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(80) NOT NULL UNIQUE,
    name_th VARCHAR(160) NOT NULL,
    version VARCHAR(32) NOT NULL,
    tradition VARCHAR(80) NOT NULL,
    source_url TEXT NOT NULL,
    conflict_policy JSONB NOT NULL DEFAULT '{"exclude_wins":true,"score_group_strategy":"highest_priority_only"}'::jsonb,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS auspicious_activities (
    code VARCHAR(64) PRIMARY KEY,
    name_th VARCHAR(120) NOT NULL,
    icon VARCHAR(16) NOT NULL,
    description_th TEXT NOT NULL DEFAULT '',
    sort_order INT NOT NULL DEFAULT 0,
    enabled BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS auspicious_rules (
    id BIGSERIAL PRIMARY KEY,
    rule_set_id BIGINT NOT NULL REFERENCES auspicious_rule_sets(id) ON DELETE RESTRICT,
    code VARCHAR(120) NOT NULL UNIQUE,
    signal_group VARCHAR(80) NOT NULL,
    name_th VARCHAR(160) NOT NULL,
    description_th TEXT NOT NULL,
    effect VARCHAR(16) NOT NULL CHECK (effect IN ('exclude', 'score')),
    score INT NOT NULL DEFAULT 0 CHECK (score BETWEEN -100 AND 100),
    priority INT NOT NULL DEFAULT 0,
    condition_json JSONB NOT NULL,
    source_note TEXT NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS auspicious_rule_activities (
    rule_id BIGINT NOT NULL REFERENCES auspicious_rules(id) ON DELETE CASCADE,
    activity_code VARCHAR(64) NOT NULL REFERENCES auspicious_activities(code) ON DELETE RESTRICT,
    PRIMARY KEY (rule_id, activity_code)
);

CREATE INDEX IF NOT EXISTS idx_auspicious_rules_active_priority
    ON auspicious_rules (rule_set_id, enabled, priority DESC);
CREATE INDEX IF NOT EXISTS idx_auspicious_rule_activities_activity
    ON auspicious_rule_activities (activity_code, rule_id);

INSERT INTO auspicious_rule_sets (code, name_th, version, tradition, source_url, conflict_policy)
VALUES (
    'isan_mahamodo_v1',
    'ฤกษ์งามยามดีแบบอีสาน',
    '1.0.0',
    'อีสาน',
    'https://www.mahamodo.com/tamnai/good_time_other_isan.aspx',
    '{"exclude_wins":true,"score_group_strategy":"highest_priority_only","notes":["วันจมชนะวันดี","ข้อห้ามเฉพาะกิจกรรมชนะคะแนนบวก","กฎดีชื่อซ้ำในกลุ่มเดียวกันไม่นับคะแนนซ้ำ"]}'::jsonb
)
ON CONFLICT (code) DO UPDATE SET
    name_th = EXCLUDED.name_th,
    version = EXCLUDED.version,
    tradition = EXCLUDED.tradition,
    source_url = EXCLUDED.source_url,
    conflict_policy = EXCLUDED.conflict_policy,
    updated_at = NOW();

INSERT INTO auspicious_activities (code, name_th, icon, description_th, sort_order) VALUES
    ('wedding', 'แต่งงาน', '💍', 'หาวันเหมาะสำหรับพิธีแต่งงานและงานมงคลสมรส', 1),
    ('housewarming', 'ขึ้นบ้านใหม่', '🏠', 'หาวันเหมาะสำหรับขึ้นบ้านใหม่หรือทำบุญบ้าน', 2),
    ('business_opening', 'เปิดร้านหรือกิจการ', '🏪', 'หาวันเหมาะสำหรับเปิดร้าน เริ่มกิจการ หรือเริ่มค้าขาย', 3)
ON CONFLICT (code) DO UPDATE SET
    name_th = EXCLUDED.name_th,
    icon = EXCLUDED.icon,
    description_th = EXCLUDED.description_th,
    sort_order = EXCLUDED.sort_order,
    enabled = TRUE;

WITH rule_set AS (
    SELECT id FROM auspicious_rule_sets WHERE code = 'isan_mahamodo_v1'
), seed(code, signal_group, name_th, description_th, effect, score, priority, condition_json, source_note) AS (
    VALUES
    (
      'isan_day_jom', 'day_jom_fu', 'วันจม',
      'เป็นวันจมตามเดือนจันทรคติ ตำราระบุให้เว้นงานมงคลแม้จะตรงกับวันดีข้ออื่น',
      'exclude', 0, 1000,
      '{"field":"weekday","operator":"equals_mapped","map_field":"lunar_month","value":{"1":"friday","2":"saturday","3":"sunday","4":"monday","5":"tuesday","6":"wednesday","7":"thursday","8":"sunday","9":"monday","10":"tuesday","11":"wednesday","12":"thursday"}}'::jsonb,
      'ตารางวันจมวันฟู; เว็บไซต์ระบุชัดว่าวันจมให้เว้นแม้ตรงวันอำฤตโชค'
    ),
    (
      'isan_day_fu', 'day_jom_fu', 'วันฟู',
      'เป็นวันฟูตามเดือนจันทรคติ เชื่อว่าเหมาะกับการเริ่มงานมงคลและความก้าวหน้า',
      'score', 12, 300,
      '{"field":"weekday","operator":"equals_mapped","map_field":"lunar_month","value":{"1":"monday","2":"tuesday","3":"wednesday","4":"thursday","5":"friday","6":"saturday","7":"sunday","8":"wednesday","9":"thursday","10":"friday","11":"saturday","12":"sunday"}}'::jsonb,
      'ตารางวันจมวันฟู'
    ),
    (
      'isan_amutachok', 'maha_chok', 'อมุตโชค',
      'ตรงดิถีอมุตโชค ซึ่งตำราจัดเป็นวันมงคลเด่น',
      'score', 18, 500,
      '{"field":"lunar_day","operator":"equals_mapped","map_field":"weekday","value":{"sunday":8,"monday":3,"tuesday":9,"wednesday":2,"thursday":4,"friday":1,"saturday":5}}'::jsonb,
      'ตารางวันอมุตโชคและแถวอมฤตโชคในตารางดิถีมหาโชคใช้ชุดตัวเลขเดียวกัน จึงเก็บเป็นสัญญาณเดียวเพื่อไม่ให้คะแนนซ้ำ'
    ),
    (
      'isan_sitthichok', 'maha_chok', 'สิทธิโชค',
      'ตรงดิถีสิทธิโชคในกลุ่มวันที่ตำราจัดว่าดี',
      'score', 14, 460,
      '{"field":"lunar_day","operator":"equals_mapped","map_field":"weekday","value":{"sunday":11,"monday":5,"tuesday":14,"wednesday":10,"thursday":9,"friday":11,"saturday":4}}'::jsonb,
      'ตารางดิถีมหาโชค'
    ),
    (
      'isan_mahasitthichok', 'maha_chok', 'มหาสิทธิโชค',
      'ตรงดิถีมหาสิทธิโชคในกลุ่มวันที่ตำราจัดว่าดี',
      'score', 16, 480,
      '{"field":"lunar_day","operator":"equals_mapped","map_field":"weekday","value":{"sunday":14,"monday":12,"tuesday":13,"wednesday":4,"thursday":7,"friday":10,"saturday":15}}'::jsonb,
      'ตารางดิถีมหาโชค'
    ),
    (
      'isan_chaichok', 'maha_chok', 'ชัยโชค',
      'ตรงดิถีชัยโชคในกลุ่มวันที่ตำราจัดว่าดี',
      'score', 14, 450,
      '{"field":"lunar_day","operator":"equals_mapped","map_field":"weekday","value":{"sunday":8,"monday":3,"tuesday":13,"wednesday":10,"thursday":4,"friday":1,"saturday":11}}'::jsonb,
      'ตารางดิถีมหาโชค; อาจตรงอมุตโชคในบางวัน ระบบเลือกคะแนนสูงสุดในกลุ่มเดียว'
    ),
    (
      'isan_rachachok', 'maha_chok', 'ราชาโชค',
      'ตรงดิถีราชาโชคในกลุ่มวันที่ตำราจัดว่าดี',
      'score', 15, 470,
      '{"field":"lunar_day","operator":"equals_mapped","map_field":"weekday","value":{"sunday":6,"monday":3,"tuesday":9,"wednesday":6,"thursday":12,"friday":1,"saturday":5}}'::jsonb,
      'ตารางดิถีมหาโชค; อาจตรงอมุตโชคในบางวัน ระบบเลือกคะแนนสูงสุดในกลุ่มเดียว'
    ),
    (
      'isan_bad_thuekthuen', 'maha_chok_bad', 'ทึกทึน', 'ตรงดิถีทึกทึนที่ตำราระบุให้ละเว้น', 'exclude', 0, 900,
      '{"field":"lunar_day","operator":"equals_mapped","map_field":"weekday","value":{"sunday":1,"monday":4,"tuesday":6,"wednesday":9,"thursday":5,"friday":3,"saturday":7}}'::jsonb, 'ตารางดิถีมหาโชค กลุ่มชั่ว-ร้าย'
    ),
    (
      'isan_bad_thorathuek', 'maha_chok_bad', 'ทรธึก', 'ตรงดิถีทรธึกที่ตำราระบุให้ละเว้น', 'exclude', 0, 900,
      '{"field":"lunar_day","operator":"equals_mapped","map_field":"weekday","value":{"sunday":4,"monday":6,"tuesday":1,"wednesday":3,"thursday":8,"friday":7,"saturday":1}}'::jsonb, 'ตารางดิถีมหาโชค กลุ่มเลว'
    ),
    (
      'isan_bad_yamakhant', 'maha_chok_bad', 'ยมขันธ์', 'ตรงดิถียมขันธ์ที่ตำราระบุให้ละเว้น', 'exclude', 0, 900,
      '{"field":"lunar_day","operator":"equals_mapped","map_field":"weekday","value":{"sunday":12,"monday":11,"tuesday":7,"wednesday":13,"thursday":6,"friday":8,"saturday":9}}'::jsonb, 'ตารางดิถีมหาโชค กลุ่มเลว'
    ),
    (
      'isan_bad_atanirot', 'maha_chok_bad', 'อัตนิโรจน์', 'ตรงดิถีอัตนิโรจน์ที่ตำราระบุให้ละเว้น', 'exclude', 0, 900,
      '{"field":"lunar_day","operator":"equals_mapped","map_field":"weekday","value":{"sunday":4,"monday":6,"tuesday":1,"wednesday":3,"thursday":3,"friday":9,"saturday":1}}'::jsonb, 'ตารางดิถีมหาโชค กลุ่มเลว'
    ),
    (
      'isan_bad_thinakan', 'maha_chok_bad', 'ทินกาล', 'ตรงดิถีทินกาลที่ตำราระบุให้ละเว้น', 'exclude', 0, 900,
      '{"field":"lunar_day","operator":"equals_mapped","map_field":"weekday","value":{"sunday":1,"monday":2,"tuesday":10,"wednesday":7,"thursday":1,"friday":6,"saturday":6}}'::jsonb, 'ตารางดิถีมหาโชค กลุ่มเลว'
    ),
    (
      'isan_bad_thinsun', 'maha_chok_bad', 'ทินสูญ', 'ตรงดิถีทินสูญที่ตำราระบุให้ละเว้น', 'exclude', 0, 900,
      '{"field":"lunar_day","operator":"equals_mapped","map_field":"weekday","value":{"sunday":12,"monday":10,"tuesday":15,"wednesday":8,"thursday":5,"friday":7,"saturday":8}}'::jsonb, 'ตารางดิถีมหาโชค กลุ่มเลว'
    ),
    (
      'isan_bad_kalachok', 'maha_chok_bad', 'กาฬโชค', 'ตรงดิถีกาฬโชคที่ตำราระบุให้ละเว้น', 'exclude', 0, 900,
      '{"field":"lunar_day","operator":"equals_mapped","map_field":"weekday","value":{"sunday":4,"monday":6,"tuesday":1,"wednesday":3,"thursday":8,"friday":9,"saturday":10}}'::jsonb, 'ตารางดิถีมหาโชค กลุ่มเลว'
    ),
    (
      'isan_bad_kalasun', 'maha_chok_bad', 'กาลสูญ', 'ตรงดิถีกาลสูญที่ตำราระบุให้ละเว้น', 'exclude', 0, 900,
      '{"field":"lunar_day","operator":"equals_mapped","map_field":"weekday","value":{"sunday":4,"monday":2,"tuesday":7,"wednesday":5,"thursday":8,"friday":3,"saturday":6}}'::jsonb, 'ตารางดิถีมหาโชค กลุ่มเลว'
    ),
    (
      'isan_bad_kalathan', 'maha_chok_bad', 'กาลทัณฑ์', 'ตรงดิถีกาลทัณฑ์ที่ตำราระบุให้ละเว้น', 'exclude', 0, 900,
      '{"field":"lunar_day","operator":"equals_mapped","map_field":"weekday","value":{"sunday":12,"monday":11,"tuesday":10,"wednesday":9,"thursday":8,"friday":7,"saturday":1}}'::jsonb, 'ตารางดิถีมหาโชค กลุ่มเลว'
    ),
    (
      'isan_bad_lokawinat', 'maha_chok_bad', 'โลกวินาส', 'ตรงดิถีโลกวินาสที่ตำราระบุให้ละเว้น', 'exclude', 0, 900,
      '{"field":"lunar_day","operator":"equals_mapped","map_field":"weekday","value":{"sunday":4,"monday":6,"tuesday":10,"wednesday":9,"thursday":8,"friday":9,"saturday":1}}'::jsonb, 'ตารางดิถีมหาโชค กลุ่มเลว'
    ),
    (
      'isan_bad_winat', 'maha_chok_bad', 'วินาสส์', 'ตรงดิถีวินาสส์ที่ตำราระบุให้ละเว้น', 'exclude', 0, 900,
      '{"field":"lunar_day","operator":"equals_mapped","map_field":"weekday","value":{"sunday":4,"monday":8,"tuesday":6,"wednesday":4,"thursday":8,"friday":8,"saturday":9}}'::jsonb, 'ตารางดิถีมหาโชค กลุ่มเลว'
    ),
    (
      'isan_bad_phila', 'maha_chok_bad', 'พิลา', 'ตรงดิถีพิลาในกลุ่มวันที่ตำราระบุให้ละเว้น', 'exclude', 0, 900,
      '{"field":"lunar_day","operator":"equals_mapped","map_field":"weekday","value":{"sunday":6,"monday":10,"tuesday":8,"wednesday":7,"thursday":2,"friday":9,"saturday":12}}'::jsonb, 'ตารางดิถีมหาโชค กลุ่มเลว'
    ),
    (
      'isan_bad_mrityu', 'maha_chok_bad', 'มฤตยู', 'ตรงดิถีมฤตยูที่ตำราระบุให้ละเว้น', 'exclude', 0, 900,
      '{"field":"lunar_day","operator":"equals_mapped","map_field":"weekday","value":{"sunday":9,"monday":1,"tuesday":10,"wednesday":9,"thursday":8,"friday":7,"saturday":6}}'::jsonb, 'ตารางดิถีมหาโชค กลุ่มเลว'
    ),
    (
      'isan_bad_blind_day', 'maha_chok_bad', 'วันบอด', 'ตรงวันบอดที่ตำราระบุให้ละเว้น', 'exclude', 0, 900,
      '{"field":"lunar_day","operator":"equals_mapped","map_field":"weekday","value":{"sunday":7,"monday":8,"tuesday":4,"wednesday":7,"thursday":1,"friday":14,"saturday":14}}'::jsonb, 'ตารางดิถีมหาโชค กลุ่มเลว'
    ),
    (
      'isan_bad_kalatin', 'maha_chok_bad', 'กาลทีน', 'ตรงดิถีกาลทีนที่ตำราระบุให้ละเว้น', 'exclude', 0, 900,
      '{"field":"lunar_day","operator":"equals_mapped","map_field":"weekday","value":{"sunday":5,"monday":6,"tuesday":10,"wednesday":8,"thursday":11,"friday":5,"saturday":7}}'::jsonb, 'ตารางดิถีมหาโชค กลุ่มเลว'
    ),
    (
      'isan_mahasun', 'mahasun', 'ดิถีมหาสูญ',
      'ตรงดิถีมหาสูญตามคู่เดือนและค่ำ ตำราระบุให้เว้นงานมงคลแม้จะตรงเงื่อนไขดีข้ออื่น',
      'exclude', 0, 950,
      '{"field":"lunar_day","operator":"equals_mapped","map_field":"lunar_month","value":{"1":2,"2":12,"3":4,"4":2,"5":6,"6":4,"7":8,"8":6,"9":10,"10":8,"11":12,"12":10}}'::jsonb,
      'ตารางดิถีมหาสูญ: เดือน 6/3=4 ค่ำ, 7/10=8 ค่ำ, 8/5=6 ค่ำ, 11/2=12 ค่ำ, 9/12=10 ค่ำ, 4/1=2 ค่ำ'
    ),
    (
      'isan_monk_avoid_14', 'ceremony_specific_block', 'สงฆ์ 14',
      'เมื่อมีพิธีสงฆ์ ตำราระบุให้เลี่ยงวันขึ้นหรือแรม 14 ค่ำ',
      'exclude', 0, 980,
      '{"all":[{"field":"monk_ceremony","operator":"eq","value":true},{"field":"lunar_day","operator":"eq","value":14}]}'::jsonb,
      'หัวข้อดิถีฤกษ์ห้ามกระทำมงคล: สงฆ์ 14'
    ),
    (
      'isan_marriage_avoid_7', 'activity_specific_block', 'สมรส 7',
      'ตำราระบุให้เลี่ยงพิธีสมรสในวันขึ้นหรือแรม 7 ค่ำ',
      'exclude', 0, 990,
      '{"field":"lunar_day","operator":"eq","value":7}'::jsonb,
      'หัวข้อดิถีฤกษ์ห้ามกระทำมงคล: สมรส 7; ขึ้น 7 ค่ำตรงวันศุกร์ย้ำว่าห้ามเด็ดขาด'
    ),
    (
      'isan_punnaditthi_business', 'activity_ditthi', 'ปุณณดิถี',
      'วันพฤหัสบดีที่ตรง 5, 10 หรือ 15 ค่ำ ตำราระบุว่าเหมาะกับการค้าและการเพาะปลูก',
      'score', 20, 520,
      '{"all":[{"field":"weekday","operator":"eq","value":"thursday"},{"field":"lunar_day","operator":"in","value":[5,10,15]}]}'::jsonb,
      'ตารางวันดีถีทั้งห้า: ปุณณดิถี'
    ),
    (
      'isan_nandaditthi_house', 'activity_ditthi', 'นันทดีถี',
      'วันศุกร์ที่ตรง 1, 6 หรือ 11 ค่ำ ตำราระบุว่าเหมาะกับการปลูกเรือน',
      'score', 20, 520,
      '{"all":[{"field":"weekday","operator":"eq","value":"friday"},{"field":"lunar_day","operator":"in","value":[1,6,11]}]}'::jsonb,
      'ตารางวันดีถีทั้งห้า: นันทดีถี'
    )
)
INSERT INTO auspicious_rules (
    rule_set_id, code, signal_group, name_th, description_th,
    effect, score, priority, condition_json, source_note
)
SELECT rule_set.id, seed.code, seed.signal_group, seed.name_th, seed.description_th,
       seed.effect, seed.score, seed.priority, seed.condition_json, seed.source_note
FROM rule_set CROSS JOIN seed
ON CONFLICT (code) DO UPDATE SET
    signal_group = EXCLUDED.signal_group,
    name_th = EXCLUDED.name_th,
    description_th = EXCLUDED.description_th,
    effect = EXCLUDED.effect,
    score = EXCLUDED.score,
    priority = EXCLUDED.priority,
    condition_json = EXCLUDED.condition_json,
    source_note = EXCLUDED.source_note,
    enabled = TRUE,
    updated_at = NOW();

INSERT INTO auspicious_rule_activities (rule_id, activity_code)
SELECT r.id, mapping.activity_code
FROM (VALUES
    ('isan_marriage_avoid_7', 'wedding'),
    ('isan_punnaditthi_business', 'business_opening'),
    ('isan_nandaditthi_house', 'housewarming')
) AS mapping(rule_code, activity_code)
JOIN auspicious_rules r ON r.code = mapping.rule_code
ON CONFLICT DO NOTHING;

COMMIT;
