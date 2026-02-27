-- 003 reconciles daily_logs with the current Go model contract:
-- - symptom_ids must stay nullable
-- - flow must not be constrained by a table-level CHECK
-- SQLite cannot drop constraints in-place, so we rebuild the table
-- in a single transaction and copy all rows before swapping tables.

-- 1) Create replacement table with the target schema.
CREATE TABLE daily_logs_new (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  date DATE NOT NULL,
  is_period BOOLEAN NOT NULL DEFAULT 0,
  flow TEXT NOT NULL DEFAULT 'none',
  symptom_ids TEXT,
  notes TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  UNIQUE(user_id, date)
);

-- 2) Copy existing data as-is to avoid loss and normalize only NULL/blank flow
-- to the default so the NOT NULL target column is satisfied.
INSERT INTO daily_logs_new (id, user_id, date, is_period, flow, symptom_ids, notes, created_at, updated_at)
SELECT
  id,
  user_id,
  date,
  is_period,
  CASE
    WHEN flow IS NULL OR trim(flow) = '' THEN 'none'
    ELSE flow
  END,
  symptom_ids,
  notes,
  COALESCE(created_at, CURRENT_TIMESTAMP),
  COALESCE(updated_at, COALESCE(created_at, CURRENT_TIMESTAMP))
FROM daily_logs;

-- 3) Swap old and new tables only after a successful copy.
DROP TABLE daily_logs;
ALTER TABLE daily_logs_new RENAME TO daily_logs;

-- 4) Restore supporting indexes expected by query paths.
CREATE INDEX IF NOT EXISTS idx_daily_logs_user_id ON daily_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_daily_logs_date ON daily_logs(date);

-- 5) Keep AUTOINCREMENT sequence aligned with migrated IDs.
INSERT OR REPLACE INTO sqlite_sequence(name, seq)
SELECT 'daily_logs', COALESCE(MAX(id), 0) FROM daily_logs;
