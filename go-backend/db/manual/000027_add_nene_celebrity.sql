-- Add the celebrity example requested for the landing-page marquee.
BEGIN;

UPDATE naming_examples
SET
  initial = 'น',
  color = '#d951a5',
  avatar_url = '/uploads/avatars/nene.jpeg',
  is_celebrity = true,
  sort_order = 9
WHERE name = 'เนเน่';

INSERT INTO naming_examples (name, initial, color, avatar_url, is_celebrity, sort_order)
SELECT 'เนเน่', 'น', '#d951a5', '/uploads/avatars/nene.jpeg', true, 9
WHERE NOT EXISTS (
  SELECT 1 FROM naming_examples WHERE name = 'เนเน่'
);

COMMIT;
