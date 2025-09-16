-- DROP trigger if exist
DROP TRIGGER IF EXISTS set_user_updated_at on users;

-- DROP trigger function
DROP FUNCTION IF EXISTS update_user_updated_at_column;

-- Drop index
DROP INDEX IF EXISTS idx_user_status;

DROP INDEX IF EXISTS idx_user_level;

DROP INDEX IF EXISTS idx_user_updated_at;

DROP INDEX IF EXISTS idx_user_deleted_at;

DROP INDEX IF EXISTS idx_user_email_status;

DROP TABLE IF EXISTS users;
