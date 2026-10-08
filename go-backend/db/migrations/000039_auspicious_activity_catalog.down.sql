BEGIN;
DELETE FROM auspicious_activities WHERE code IN ('name_change', 'vehicle_delivery', 'childbirth', 'ordination');
ALTER TABLE auspicious_activities DROP COLUMN IF EXISTS safety_note_th;
ALTER TABLE auspicious_activities DROP COLUMN IF EXISTS rule_scope_note_th;
COMMIT;
