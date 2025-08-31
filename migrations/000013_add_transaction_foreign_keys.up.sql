-- Migration to add foreign key constraints to transactions table for proper relationships

-- Add foreign key constraint for smart_cheque_id
ALTER TABLE transactions 
ADD CONSTRAINT fk_transactions_smart_cheque_id 
FOREIGN KEY (smart_cheque_id) REFERENCES smart_cheques(id) 
ON DELETE SET NULL;

-- Add foreign key constraint for milestone_id
ALTER TABLE transactions 
ADD CONSTRAINT fk_transactions_milestone_id 
FOREIGN KEY (milestone_id) REFERENCES milestones(id) 
ON DELETE SET NULL;