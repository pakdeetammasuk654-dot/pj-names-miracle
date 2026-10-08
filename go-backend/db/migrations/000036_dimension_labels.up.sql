CREATE TABLE IF NOT EXISTS dimension_labels (
    id SERIAL PRIMARY KEY,
    dimension VARCHAR(32) NOT NULL,
    min_score NUMERIC(5,2) NOT NULL,
    label VARCHAR(128) NOT NULL,
    description TEXT DEFAULT '',
    sort_order INT DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_dimension_labels_dim_score ON dimension_labels (dimension, min_score DESC);

INSERT INTO dimension_labels (dimension, min_score, label, description, sort_order) VALUES
('meaning', 90.0, 'โดดเด่นระดับสูง', 'คำแปลมีความหมายลึกซึ้ง ทรงพลังบารมีอย่างยิ่ง', 1),
('meaning', 80.0, 'เด่นมาก มงคลลึกซึ้ง', 'ความหมายมงคลรอบด้าน เกื้อหนุนชีวิต', 2),
('meaning', 70.0, 'ความหมายดี มงคล', 'ความหมายดีตามหลักสิริมงคล', 3),
('meaning', 0.0, 'ความหมายทั่วไป', 'ความหมายทั่วไปตามพจนานุกรม', 4),

('numerology', 90.0, 'คู่มงคลชั้นสูง มหาบารมี', 'คู่เลขมงคลเกรดสูงสุด ไร้ตำหนิ', 1),
('numerology', 80.0, 'คู่มงคลพลังดีเยี่ยม', 'คู่เลขมงคลส่งเสริมความสำเร็จ', 2),
('numerology', 70.0, 'คู่มงคลเกณฑ์ดี', 'เลขศาสตร์สมดุลตามเกณฑ์มงคล', 3),
('numerology', 0.0, 'คู่เลขพลังปานกลาง', 'คู่เลขมาตรฐานทั่วไป', 4),

('sound', 90.0, 'จังหวะเสียงไพเราะก้องกังวาน', 'เสียงกังวาน ทรงพลัง ลื่นไหลเป็นจังหวะทองคำ', 1),
('sound', 80.0, 'เสียงสง่างาม มีน้ำหนัก', 'เสียงมีความหนักแน่นและไพเราะน่าฟัง', 2),
('sound', 70.0, 'ไพเราะ คล่องตัว', 'ออกเสียงง่าย ลื่นไหล ไม่สะดุด', 3),
('sound', 0.0, 'เสียงยาว ลุ่มลึก', 'จังหวะเสียงทอดยาว นุ่มนวล', 4),

('resonance', 88.0, 'สมดุลรอบด้าน เอกลักษณ์สูง', 'โครงสร้างรากศัพท์สมบูรณ์แบบ ทรงเอกลักษณ์', 1),
('resonance', 75.0, 'กลมกลืน ทรงเอกลักษณ์', 'รากคำกลมกลืนกับนามสกุลและโครงสร้าง', 2),
('resonance', 0.0, 'เอกลักษณ์เฉพาะตัว', 'โครงสร้างคำเฉพาะตัว โดดเด่น', 3);
