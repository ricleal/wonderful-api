-- Add composite index for efficient cursor-based pagination
-- This index supports ORDER BY registration DESC, id DESC queries
-- and enables efficient cursor-based pagination without subqueries
CREATE INDEX index_users_on_registration_id_desc ON users(registration DESC, id DESC);

-- Drop the old single-column index since it's redundant with the composite index
DROP INDEX index_users_on_registration;
