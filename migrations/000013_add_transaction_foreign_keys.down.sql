-- Migration to remove foreign key constraints from transactions table

-- Remove foreign key constraint for smart_cheque_id
ALTER TABLE transactions 
DROP CONSTRAINT IF EXISTS fk_transactions_smart_cheque_id;

-- Remove foreign key constraint for milestone_id
ALTER TABLE transactions 
DROP CONSTRAINT IF EXISTS fk_transactions_milestone_id;