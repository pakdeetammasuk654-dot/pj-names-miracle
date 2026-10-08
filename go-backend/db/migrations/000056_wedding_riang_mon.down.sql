BEGIN;

DELETE FROM auspicious_rule_activities
WHERE activity_code = 'wedding'
  AND rule_id IN (
      SELECT id FROM auspicious_rules
      WHERE code IN (
          'thai_wedding_riang_mon',
          'thai_kala_yok_athibodi',
          'thai_kala_yok_thongchai',
          'thai_kala_yok_ubat',
          'thai_kala_yok_lokawinat'
      )
  );

DELETE FROM auspicious_rules
WHERE code = 'thai_wedding_riang_mon';

DELETE FROM auspicious_rule_sets
WHERE code = 'thai_wedding_tradition_v1';

UPDATE auspicious_rules
SET condition_json = '{"field":"lunar_day","operator":"eq","value":7}'::jsonb,
    description_th = 'ตำราระบุให้เลี่ยงพิธีสมรสในวันขึ้นหรือแรม 7 ค่ำ',
    updated_at = NOW()
WHERE code = 'isan_marriage_avoid_7';

COMMIT;
