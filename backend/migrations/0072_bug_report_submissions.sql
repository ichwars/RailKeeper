CREATE TABLE bug_report_submissions (
  report_id TEXT PRIMARY KEY,
  actor_user_id TEXT NOT NULL,
  state TEXT NOT NULL CHECK(state IN ('reserved', 'issue_created', 'asset_uploaded', 'complete', 'partial')),
  github_issue_number INTEGER,
  github_issue_url TEXT NOT NULL DEFAULT '',
  asset_url TEXT NOT NULL DEFAULT '',
  asset_sha TEXT NOT NULL DEFAULT '',
  error_category TEXT NOT NULL DEFAULT '' CHECK(length(error_category) <= 64),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE INDEX idx_bug_report_submissions_actor_created
ON bug_report_submissions(actor_user_id, created_at);
