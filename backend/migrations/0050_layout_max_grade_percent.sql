ALTER TABLE layouts ADD COLUMN max_grade_percent REAL
    CHECK(max_grade_percent IS NULL OR (max_grade_percent > 0 AND max_grade_percent <= 100));
