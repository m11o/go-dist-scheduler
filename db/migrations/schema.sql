-- tasks table
-- The payload column stores HTTPRequestInfo as JSON with the following structure:
-- {
--   "url": "string",
--   "method": "string",
--   "headers": {"key": "value", ...},
--   "body": "base64-encoded string"
-- }
CREATE TABLE IF NOT EXISTS tasks (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    cron_expression VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    status INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index for querying active tasks
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);

-- Index for querying by created_at
CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at);
