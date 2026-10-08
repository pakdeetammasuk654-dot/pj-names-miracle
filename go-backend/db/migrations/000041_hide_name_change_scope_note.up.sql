BEGIN;

UPDATE auspicious_activities
SET rule_scope_note_th = ''
WHERE code = 'name_change';

COMMIT;
