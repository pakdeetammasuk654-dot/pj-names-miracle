BEGIN;

ALTER TABLE auspicious_rules
    DROP COLUMN IF EXISTS meaning_th;

COMMIT;
