ALTER TABLE articles
  ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

CREATE INDEX IF NOT EXISTS idx_articles_updated_at
  ON articles (updated_at DESC);
