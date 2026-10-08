CREATE INDEX IF NOT EXISTS idx_names_miracle_sat_sha_sum
  ON names_miracle (sat_sum, sha_sum);
