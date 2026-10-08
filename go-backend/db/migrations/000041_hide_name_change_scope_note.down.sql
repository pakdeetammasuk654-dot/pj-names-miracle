BEGIN;

UPDATE auspicious_activities
SET rule_scope_note_th = 'ขณะนี้ใช้กฎวันมงคลทั่วไป เพราะตำราอีสานชุดแรกไม่ได้ระบุกฎเฉพาะการเปลี่ยนชื่อ'
WHERE code = 'name_change';

COMMIT;
