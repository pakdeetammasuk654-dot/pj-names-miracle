-- normalize_thai_phonetic: strips การันต์ and unifies consonant classes
-- so pg_trgm similarity can match phonetically equivalent Thai names.
CREATE OR REPLACE FUNCTION normalize_thai_phonetic(input text) RETURNS text
LANGUAGE sql IMMUTABLE STRICT PARALLEL SAFE AS $$
  SELECT translate(
    regexp_replace(input, '[ก-ฮ]์', '', 'g'),
    'ณศษฑฐภธ',
    'นสสททพท'
  )
$$;
