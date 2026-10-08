BEGIN;

DELETE FROM auspicious_rule_activities
WHERE activity_code IN ('brand_naming', 'important_documents', 'product_launch', 'career_start');

DELETE FROM auspicious_activities
WHERE code IN ('brand_naming', 'important_documents', 'product_launch', 'career_start');

ALTER TABLE auspicious_activities
    DROP COLUMN IF EXISTS category_code;

DROP TABLE IF EXISTS auspicious_activity_categories;

COMMIT;
