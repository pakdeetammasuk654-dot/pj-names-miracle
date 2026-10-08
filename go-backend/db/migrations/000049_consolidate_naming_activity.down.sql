BEGIN;

UPDATE auspicious_activities
SET name_th = 'เปลี่ยนชื่อ ตั้งชื่อสกุล หรือยื่นเอกสารสำคัญ',
    description_th = 'เลือกวันเริ่มใช้ชื่อใหม่ เปลี่ยนชื่อในเอกสารราชการ หรือดำเนินเรื่องสำคัญ',
    rule_scope_note_th = 'ใช้กฎทั่วไป พร้อมปุณณดิถีซึ่งตำราระบุว่าเป็นดิถีแห่งความบริบูรณ์ เหมาะแก่การตั้งชื่อและเริ่มชีวิตใหม่',
    safety_note_th = 'ควรตรวจวันทำการ เวลารับเรื่อง และข้อกำหนดของหน่วยงานก่อนเลือกวัน',
    sort_order = 1,
    enabled = TRUE
WHERE code = 'name_change';

UPDATE auspicious_activities
SET enabled = TRUE
WHERE code = 'brand_naming';

COMMIT;
