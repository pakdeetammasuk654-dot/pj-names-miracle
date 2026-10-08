BEGIN;

UPDATE auspicious_activities
SET rule_scope_note_th = ''
WHERE code = 'brand_naming';

DELETE FROM auspicious_rule_activities
WHERE activity_code = 'brand_naming'
  AND rule_id IN (SELECT id FROM auspicious_rules WHERE code = 'isan_punnaditthi_business');

COMMIT;
