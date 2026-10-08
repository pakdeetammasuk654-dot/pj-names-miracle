DROP INDEX IF EXISTS idx_names_miracle_phonetic_trgm;
DROP INDEX IF EXISTS idx_names_miracle_thname_trgm;

-- pg_trgm may be shared by other features, so the down migration intentionally
-- leaves the extension installed.
