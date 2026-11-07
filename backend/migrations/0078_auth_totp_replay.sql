ALTER TABLE users ADD COLUMN two_factor_last_step INTEGER
  CHECK (two_factor_last_step IS NULL OR two_factor_last_step >= 0);
