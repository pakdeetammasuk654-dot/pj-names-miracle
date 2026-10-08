BEGIN;

UPDATE auspicious_rule_sets
SET name_th = 'ฤกษ์วันมงคลตามตำราอีสาน',
    updated_at = NOW()
WHERE code = 'isan_mahamodo_v1';

COMMIT;
