BEGIN;

DELETE FROM auspicious_rule_activities
WHERE activity_code = 'negotiation';

DELETE FROM auspicious_activities
WHERE code = 'negotiation';

DELETE FROM auspicious_rules
WHERE code IN (
    'thai_kala_yok_athibodi',
    'thai_kala_yok_thongchai',
    'thai_kala_yok_ubat',
    'thai_kala_yok_lokawinat'
);

DELETE FROM auspicious_rule_sets
WHERE code = 'thai_kala_yok_v1';

COMMIT;
