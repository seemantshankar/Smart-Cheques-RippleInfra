-- Migration Down: Remove transaction management tables
-- Description: Remove tables and objects created for transaction queue management

-- Drop views first (dependent objects)
DROP VIEW IF EXISTS hourly_transaction_stats;
DROP VIEW IF EXISTS batch_stats_view;
DROP VIEW IF EXISTS transaction_stats_view;

-- Drop triggers
DROP TRIGGER IF EXISTS update_transaction_batches_updated_at ON transaction_batches;
DROP TRIGGER IF EXISTS update_transactions_updated_at ON transactions;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop foreign key constraint
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS fk_transactions_batch_id;

-- Drop indexes
DROP INDEX IF EXISTS idx_transaction_batches_created_at;
DROP INDEX IF EXISTS idx_transaction_batches_priority_created;
DROP INDEX IF EXISTS idx_transaction_batches_status;

DROP INDEX IF EXISTS idx_transactions_expires_at;
DROP INDEX IF EXISTS idx_transactions_transaction_hash;
DROP INDEX IF EXISTS idx_transactions_milestone_id;
DROP INDEX IF EXISTS idx_transactions_smart_cheque_id;
DROP INDEX IF EXISTS idx_transactions_priority_created;
DROP INDEX IF EXISTS idx_transactions_type;
DROP INDEX IF EXISTS idx_transactions_user_id;
DROP INDEX IF EXISTS idx_transactions_enterprise_id;
DROP INDEX IF EXISTS idx_transactions_batch_id;
DROP INDEX IF EXISTS idx_transactions_status;

-- Drop tables
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS transaction_batches;