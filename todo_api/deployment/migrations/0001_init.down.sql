DROP INDEX IF EXISTS idx_tasks_completed;
DROP INDEX IF EXISTS idx_tasks_created_at;
DROP INDEX IF EXISTS idx_outbox_sent;
DROP INDEX IF EXISTS idx_outbox_created_at;
DROP TABLE IF EXISTS outbox;
DROP TABLE IF EXISTS tasks;