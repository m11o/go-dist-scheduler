DROP TRIGGER IF EXISTS set_updated_at ON tasks;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS tasks;

-- Note: We intentionally do not drop the uuid-ossp extension here
-- as it may be used by other tables or future migrations.
