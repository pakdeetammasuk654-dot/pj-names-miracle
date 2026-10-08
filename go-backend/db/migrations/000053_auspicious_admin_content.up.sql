BEGIN;

CREATE TABLE IF NOT EXISTS auspicious_texts (
    key VARCHAR(160) PRIMARY KEY,
    group_code VARCHAR(64) NOT NULL,
    label_th VARCHAR(160) NOT NULL,
    value_th TEXT NOT NULL,
    help_th TEXT NOT NULL DEFAULT '',
    sort_order INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_auspicious_texts_group_sort
    ON auspicious_texts (group_code, sort_order, key);

INSERT INTO auspicious_texts (key, group_code, label_th, value_th, help_th, sort_order) VALUES
('taksa.role.boriwan.name', 'taksa_roles', 'ชื่อภูมิบริวาร', 'บริวาร', '', 10),
('taksa.role.boriwan.meaning', 'taksa_roles', 'ความหมายภูมิบริวาร', 'ภูมิบริวารหมายถึงมิตรภาพ ผู้ช่วยเหลือ ผู้ร่วมงาน และความสามัคคี', '', 11),
('taksa.role.boriwan.description', 'taksa_roles', 'คำอธิบายวงล้อภูมิบริวาร', 'บริวาร มิตรภาพ ผู้ร่วมงาน ลูกน้อง และคนรอบข้าง', '', 12),
('taksa.role.ayu.name', 'taksa_roles', 'ชื่อภูมิอายุ', 'อายุ', '', 20),
('taksa.role.ayu.meaning', 'taksa_roles', 'ความหมายภูมิอายุ', 'ภูมิอายุหมายถึงสุขภาพ ความยั่งยืน ความราบรื่น และอายุยืนยาว', '', 21),
('taksa.role.ayu.description', 'taksa_roles', 'คำอธิบายวงล้อภูมิอายุ', 'อายุ สุขภาพ พลานามัย ความเป็นอยู่ และวิถีชีวิต', '', 22),
('taksa.role.dech.name', 'taksa_roles', 'ชื่อภูมิเดช', 'เดช', '', 30),
('taksa.role.dech.meaning', 'taksa_roles', 'ความหมายภูมิเดช', 'ภูมิเดชหมายถึงอำนาจ บารมี ความเข้มแข็ง และความเป็นผู้นำ', '', 31),
('taksa.role.dech.description', 'taksa_roles', 'คำอธิบายวงล้อภูมิเดช', 'เดช อำนาจ บารมี เกียรติยศ และความก้าวหน้า', '', 32),
('taksa.role.sri.name', 'taksa_roles', 'ชื่อภูมิศรี', 'ศรี', '', 40),
('taksa.role.sri.meaning', 'taksa_roles', 'ความหมายภูมิศรี', 'ภูมิศรีหมายถึงความดีงาม สิริมงคล โชคลาภ และความสำเร็จราบรื่น', '', 41),
('taksa.role.sri.description', 'taksa_roles', 'คำอธิบายวงล้อภูมิศรี', 'ศรี สิริมงคล โชคลาภ ทรัพย์สิน เมตตามหานิยม', '', 42),
('taksa.role.mula.name', 'taksa_roles', 'ชื่อภูมิมูละ', 'มูละ', '', 50),
('taksa.role.mula.meaning', 'taksa_roles', 'ความหมายภูมิมูละ', 'ภูมิมูละหมายถึงความมั่นคง ทรัพย์สิน มรดก และหลักฐานมั่นคง', '', 51),
('taksa.role.mula.description', 'taksa_roles', 'คำอธิบายวงล้อภูมิมูละ', 'มูละ ความมั่นคง หลักฐาน ทรัพย์สิน ที่อยู่อาศัย', '', 52),
('taksa.role.utsaha.name', 'taksa_roles', 'ชื่อภูมิอุตสาหะ', 'อุตสาหะ', '', 60),
('taksa.role.utsaha.meaning', 'taksa_roles', 'ความหมายภูมิอุตสาหะ', 'ภูมิอุตสาหะหมายถึงความขยัน มุ่งมั่น มานะพยายามสู่ความสำเร็จ', '', 61),
('taksa.role.utsaha.description', 'taksa_roles', 'คำอธิบายวงล้อภูมิอุตสาหะ', 'อุตสาหะ ความขยัน มุ่งมั่น ความเพียรพยายาม', '', 62),
('taksa.role.montri.name', 'taksa_roles', 'ชื่อภูมิมนตรี', 'มนตรี', '', 70),
('taksa.role.montri.meaning', 'taksa_roles', 'ความหมายภูมิมนตรี', 'ภูมิมนตรีหมายถึงการอุปถัมภ์ค้ำชู ที่พึ่งพา และความเมตตาเอ็นดูจากผู้ใหญ่', '', 71),
('taksa.role.montri.description', 'taksa_roles', 'คำอธิบายวงล้อภูมิมนตรี', 'มนตรี ผู้ใหญ่อุปถัมภ์ เมตตาช่วยเหลือ ราบรื่น', '', 72),
('taksa.role.kalakini.name', 'taksa_roles', 'ชื่อภูมิกาลกิณี', 'กาลกิณี', '', 80),
('taksa.role.kalakini.meaning', 'taksa_roles', 'ความหมายภูมิกาลกิณี', 'กาลกิณีหมายถึงอุปสรรค สิ่งติดขัด หรือความไม่ราบรื่น โบราณมักให้หลีกเลี่ยง', '', 81),
('taksa.role.kalakini.description', 'taksa_roles', 'คำอธิบายวงล้อภูมิกาลกิณี', 'กาลกิณี อุปสรรค ความขัดแย้ง สิ่งที่ควรหลีกเลี่ยง', '', 82),
('taksa.template.progressed_role_name', 'taksa_templates', 'ชื่อภูมิบนวงล้อทักษาจร', '{role}จร', 'ตัวแปรที่ใช้ได้: {role}', 100),
('taksa.template.progressed_reason_name', 'taksa_templates', 'หัวข้อเหตุผลทักษาจร', 'วัน{role}จร (อายุย่าง {age} ปี)', 'ตัวแปรที่ใช้ได้: {role}, {age}', 110),
('taksa.template.progressed_explanation', 'taksa_templates', 'คำอธิบายทักษาจรทั่วไป', '{date} เป็นวัน{role}จรสำหรับผู้เกิด{birth_day} (อายุย่าง {age} ปี) ตามหลักทักษาจร', 'ตัวแปรที่ใช้ได้: {date}, {role}, {birth_day}, {age}', 120),
('taksa.template.progressed_kalakini', 'taksa_templates', 'คำอธิบายกาลกิณีจร', '{date} เป็นกาลกิณีจรประจำปีสำหรับผู้เกิด{birth_day} (อายุย่าง {age} ปี) ตามหลักทักษาจร ควรหลีกเลี่ยงการเริ่มการมงคล', 'ตัวแปรที่ใช้ได้: {date}, {birth_day}, {age}', 130),
('taksa.template.progressed_center_kalakini', 'taksa_templates', 'คำอธิบายกาลกิณีเมื่อตกตากลาง', '{date} เป็นกาลกิณีจรสำหรับผู้เกิด{birth_day} (อายุย่าง {age} ปี ตกภูมิกลาง) ตามหลักทักษาจร ควรหลีกเลี่ยงการเริ่มการมงคล', 'ตัวแปรที่ใช้ได้: {date}, {birth_day}, {age}', 140),
('taksa.template.natal_reason_name', 'taksa_templates', 'หัวข้อเหตุผลทักษากำเนิด', 'วัน{role}ประจำวันเกิด', 'ตัวแปรที่ใช้ได้: {role}', 150),
('taksa.template.natal_explanation', 'taksa_templates', 'คำอธิบายทักษากำเนิดทั่วไป', '{date} เป็น{name}สำหรับผู้เกิด{birth_day} ตามหลักทักษาปกรณ์', 'ตัวแปรที่ใช้ได้: {date}, {name}, {birth_day}', 160),
('taksa.template.natal_kalakini', 'taksa_templates', 'คำอธิบายกาลกิณีกำเนิด', '{date} เป็นกาลกิณีสำหรับผู้เกิด{birth_day} ตามหลักทักษาปกรณ์ ควรหลีกเลี่ยงการเริ่มการมงคล', 'ตัวแปรที่ใช้ได้: {date}, {birth_day}', 170),
('search.disclaimer', 'search', 'ข้อความกำกับผลการค้นหา', '', '', 200)
ON CONFLICT (key) DO NOTHING;

COMMIT;
