BEGIN;

UPDATE auspicious_texts
SET value_th = '', updated_at = NOW()
WHERE key = 'search.disclaimer';

COMMIT;
