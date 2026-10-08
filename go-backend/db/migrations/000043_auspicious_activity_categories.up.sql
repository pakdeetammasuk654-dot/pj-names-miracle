BEGIN;

CREATE TABLE IF NOT EXISTS auspicious_activity_categories (
    code VARCHAR(64) PRIMARY KEY,
    name_th VARCHAR(120) NOT NULL,
    icon VARCHAR(16) NOT NULL,
    description_th TEXT NOT NULL DEFAULT '',
    sort_order INT NOT NULL DEFAULT 0,
    enabled BOOLEAN NOT NULL DEFAULT TRUE
);

INSERT INTO auspicious_activity_categories (code, name_th, icon, description_th, sort_order, enabled) VALUES
    ('naming_documents', 'ชื่อและเอกสาร', '✍️', 'การตั้งชื่อ ประกาศใช้ชื่อ และดำเนินเอกสารสำคัญ', 1, TRUE),
    ('work_business', 'งานและธุรกิจ', '💼', 'การเริ่มงาน เปิดกิจการ และเปิดตัวผลงานใหม่', 2, TRUE),
    ('home_property', 'บ้านและทรัพย์สิน', '🏠', 'การขึ้นบ้าน ย้ายที่พัก และเริ่มใช้ทรัพย์สินสำคัญ', 3, TRUE),
    ('family_ceremony', 'ครอบครัวและพิธี', '🙏', 'งานมงคลของครอบครัวและพิธีทางศาสนา', 4, TRUE),
    ('travel_vehicle', 'การเดินทาง', '🚗', 'การรับรถ เริ่มใช้ยานพาหนะ และวางแผนการเดินทาง', 5, TRUE),
    ('other', 'อื่นๆ', '✦', 'กิจกรรมมงคลอื่น', 99, TRUE)
ON CONFLICT (code) DO UPDATE SET
    name_th = EXCLUDED.name_th,
    icon = EXCLUDED.icon,
    description_th = EXCLUDED.description_th,
    sort_order = EXCLUDED.sort_order,
    enabled = EXCLUDED.enabled;

ALTER TABLE auspicious_activities
    ADD COLUMN IF NOT EXISTS category_code VARCHAR(64) REFERENCES auspicious_activity_categories(code) DEFAULT 'other';

UPDATE auspicious_activities SET category_code = CASE code
    WHEN 'name_change' THEN 'naming_documents'
    WHEN 'vehicle_delivery' THEN 'travel_vehicle'
    WHEN 'childbirth' THEN 'family_ceremony'
    WHEN 'wedding' THEN 'family_ceremony'
    WHEN 'housewarming' THEN 'home_property'
    WHEN 'business_opening' THEN 'work_business'
    WHEN 'ordination' THEN 'family_ceremony'
    ELSE COALESCE(category_code, 'other')
END;

ALTER TABLE auspicious_activities
    ALTER COLUMN category_code SET DEFAULT 'other',
    ALTER COLUMN category_code SET NOT NULL;

UPDATE auspicious_activities SET sort_order = CASE code
    WHEN 'name_change' THEN 1
    WHEN 'business_opening' THEN 1
    WHEN 'housewarming' THEN 1
    WHEN 'wedding' THEN 1
    WHEN 'childbirth' THEN 2
    WHEN 'ordination' THEN 3
    WHEN 'vehicle_delivery' THEN 1
    ELSE sort_order
END;

INSERT INTO auspicious_activities (
    code, name_th, icon, description_th, rule_scope_note_th, safety_note_th,
    sort_order, enabled, category_code
) VALUES
    ('brand_naming', 'ตั้งชื่อร้าน แบรนด์ หรือบริษัท', '🏷️', 'เลือกวันประกาศใช้ชื่อธุรกิจ จดทะเบียน หรือเปิดตัวชื่อใหม่', '', '', 2, TRUE, 'naming_documents'),
    ('important_documents', 'จดทะเบียนหรือยื่นเอกสารสำคัญ', '📄', 'เลือกวันยื่นเอกสาร จดทะเบียน หรือดำเนินเรื่องสำคัญ', '', 'ควรตรวจวันทำการ เวลารับเรื่อง และข้อกำหนดของหน่วยงานก่อนเลือกวัน', 3, TRUE, 'naming_documents'),
    ('product_launch', 'เปิดตัวสินค้า เว็บไซต์ หรือร้านออนไลน์', '🚀', 'เลือกวันเปิดตัวสินค้า เว็บไซต์ แอป เพจ หรือร้านออนไลน์', 'ใช้กฎวันมงคลร่วมกับปุณณดิถีสำหรับการค้า', '', 2, TRUE, 'work_business'),
    ('career_start', 'เริ่มงานใหม่หรือรับตำแหน่ง', '💼', 'เลือกวันเริ่มงานวันแรก รับตำแหน่ง หรือเริ่มบทบาทใหม่', '', '', 3, TRUE, 'work_business')
ON CONFLICT (code) DO UPDATE SET
    name_th = EXCLUDED.name_th,
    icon = EXCLUDED.icon,
    description_th = EXCLUDED.description_th,
    rule_scope_note_th = EXCLUDED.rule_scope_note_th,
    safety_note_th = EXCLUDED.safety_note_th,
    sort_order = EXCLUDED.sort_order,
    enabled = EXCLUDED.enabled,
    category_code = EXCLUDED.category_code;

INSERT INTO auspicious_rule_activities (rule_id, activity_code)
SELECT id, 'product_launch' FROM auspicious_rules WHERE code = 'isan_punnaditthi_business'
ON CONFLICT DO NOTHING;

COMMIT;
