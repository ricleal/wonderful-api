-- Revert the composite index changes
-- Recreate the original single-column index
CREATE INDEX index_users_on_registration ON users(registration);

-- Drop the composite index
DROP INDEX index_users_on_registration_id_desc;
