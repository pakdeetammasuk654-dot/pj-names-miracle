CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_names_miracle_thname_trgm
  ON names_miracle USING gin (thname gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_names_miracle_phonetic_trgm
  ON names_miracle USING gin (normalize_thai_phonetic(thname) gin_trgm_ops);
