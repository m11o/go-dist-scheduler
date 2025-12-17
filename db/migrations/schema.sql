-- tasks table
-- The payload column stores HTTPRequestInfo as JSON with the following structure:
-- {
--   "url": "string",
--   "method": "string",
--   "headers": {"key": "value", ...},
--   "body": "base64-encoded string"
-- }
CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    cron_expression VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    status INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_checked_at TIMESTAMP NULL
);

-- Index for querying active tasks
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);

-- Index for querying by created_at
CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at);

-- jobs table
-- Represents scheduled job executions
CREATE TABLE IF NOT EXISTS jobs (
    id UUID PRIMARY KEY,
    task_id UUID NOT NULL,
    scheduled_at TIMESTAMP NOT NULL,
    started_at TIMESTAMP NULL,
    finished_at TIMESTAMP NULL,
    status INTEGER NOT NULL DEFAULT 0,
    retry_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index for querying jobs by status
CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);

-- Index for querying jobs by task_id
CREATE INDEX IF NOT EXISTS idx_jobs_task_id ON jobs(task_id);

-- Index for querying jobs by scheduled_at
CREATE INDEX IF NOT EXISTS idx_jobs_scheduled_at ON jobs(scheduled_at);
