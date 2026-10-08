BEGIN;

ALTER TABLE auspicious_activities
    ADD COLUMN IF NOT EXISTS rule_scope_note_th TEXT NOT NULL DEFAULT 'ใช้กฎมงคลทั่วไปจากชุดตำราที่เปิดใช้งาน',
    ADD COLUMN IF NOT EXISTS safety_note_th TEXT NOT NULL DEFAULT '';

UPDATE auspicious_activities SET
    sort_order = CASE code
        WHEN 'wedding' THEN 4
        WHEN 'housewarming' THEN 5
        WHEN 'business_opening' THEN 6
        ELSE sort_order
    END,
    rule_scope_note_th = CASE code
        WHEN 'wedding' THEN 'ใช้กฎทั่วไป พร้อมข้อห้ามสมรส 7 ที่ระบุเฉพาะงานแต่งงาน'
        WHEN 'housewarming' THEN 'ใช้กฎทั่วไป พร้อมนันทดีถีซึ่งตำราระบุว่าเหมาะกับการปลูกเรือน'
        WHEN 'business_opening' THEN 'ใช้กฎทั่วไป พร้อมปุณณดิถีซึ่งตำราระบุว่าเหมาะกับการค้า'
        ELSE rule_scope_note_th
    END;

INSERT INTO auspicious_activities (
    code, name_th, icon, description_th, rule_scope_note_th, safety_note_th, sort_order, enabled
) VALUES
    (
      'name_change', 'เปลี่ยนชื่อหรือตั้งชื่อสกุล', '✍️',
      'เลือกวันเริ่มใช้ชื่อใหม่ ยื่นเอกสาร หรือประกาศใช้ชื่อสกุล',
      'ขณะนี้ใช้กฎวันมงคลทั่วไป เพราะตำราอีสานชุดแรกไม่ได้ระบุกฎเฉพาะการเปลี่ยนชื่อ',
      '', 1, TRUE
    ),
    (
      'vehicle_delivery', 'รับรถหรือออกรถใหม่', '🚗',
      'เลือกวันรับรถ เริ่มใช้รถ หรือออกเดินทางครั้งแรก',
      'ขณะนี้ใช้กฎวันมงคลทั่วไป ยังไม่ใช้ตารางยามเดินทางมาตัดสินวันรับรถ',
      'ควรคำนึงถึงความพร้อมของผู้ขับขี่ ประกันภัย และสภาพรถเป็นหลัก', 2, TRUE
    ),
    (
      'childbirth', 'คลอดบุตร', '👶',
      'ใช้เป็นข้อมูลประกอบสำหรับการคลอดที่แพทย์เห็นว่าสามารถวางแผนวันได้',
      'ขณะนี้ใช้กฎวันมงคลทั่วไป ยังไม่มีตำราเฉพาะการคลอดในชุดกฎนี้',
      'วันและเวลาคลอดต้องเป็นไปตามความปลอดภัยของมารดาและทารก รวมถึงคำแนะนำของแพทย์ก่อนฤกษ์เสมอ', 3, TRUE
    ),
    (
      'ordination', 'บวชหรือทำพิธีสงฆ์', '🙏',
      'เลือกวันบวช ทำบุญ หรือประกอบพิธีที่นิมนต์พระสงฆ์',
      'ใช้กฎวันมงคลทั่วไป และควรเปิดตัวเลือก “มีพิธีสงฆ์” เพื่อตรวจสงฆ์ 14',
      '', 7, TRUE
    )
ON CONFLICT (code) DO UPDATE SET
    name_th = EXCLUDED.name_th,
    icon = EXCLUDED.icon,
    description_th = EXCLUDED.description_th,
    rule_scope_note_th = EXCLUDED.rule_scope_note_th,
    safety_note_th = EXCLUDED.safety_note_th,
    sort_order = EXCLUDED.sort_order,
    enabled = EXCLUDED.enabled;

COMMIT;
