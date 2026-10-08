BEGIN;

ALTER TABLE auspicious_rules
    ADD COLUMN IF NOT EXISTS meaning_th TEXT NOT NULL DEFAULT '';

UPDATE auspicious_rules
SET meaning_th = description_th
WHERE meaning_th = '';

UPDATE auspicious_rules
SET meaning_th = CASE code
    WHEN 'isan_day_fu' THEN 'วันฟูหมายถึงวันที่เชื่อว่าส่งเสริมความงอกงาม ความก้าวหน้า และการเริ่มต้นงานมงคล'
    WHEN 'isan_amutachok' THEN 'อมุตโชคเป็นดิถีมงคลเด่น ตามตำราถือว่าเหมาะสำหรับเริ่มต้นงานสำคัญ'
    WHEN 'isan_sitthichok' THEN 'สิทธิโชคเป็นดิถีที่สื่อถึงความสำเร็จและการบรรลุผล'
    WHEN 'isan_mahasitthichok' THEN 'มหาสิทธิโชคเป็นดิถีที่สื่อถึงความสำเร็จอย่างยิ่ง จัดเป็นวันมงคลเด่นในชุดตำรา'
    WHEN 'isan_chaichok' THEN 'ชัยโชคเป็นดิถีที่สื่อถึงชัยชนะและความสำเร็จของงาน'
    WHEN 'isan_rachachok' THEN 'ราชาโชคเป็นดิถีที่สื่อถึงเกียรติ อำนาจ และความก้าวหน้า'
    ELSE meaning_th
END
WHERE code IN (
    'isan_day_fu', 'isan_amutachok', 'isan_sitthichok',
    'isan_mahasitthichok', 'isan_chaichok', 'isan_rachachok'
);

COMMIT;
