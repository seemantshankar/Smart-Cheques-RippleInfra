-- Migration to enhance smart_cheques table with additional fields and constraints

-- Add indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_smart_cheques_payer_id ON smart_cheques(payer_id);
CREATE INDEX IF NOT EXISTS idx_smart_cheques_payee_id ON smart_cheques(payee_id);
CREATE INDEX IF NOT EXISTS idx_smart_cheques_status ON smart_cheques(status);
CREATE INDEX IF NOT EXISTS idx_smart_cheques_contract_hash ON smart_cheques(contract_hash);
CREATE INDEX IF NOT EXISTS idx_smart_cheques_created_at ON smart_cheques(created_at);

-- Add milestones column if it doesn't exist
ALTER TABLE smart_cheques 
ADD COLUMN IF NOT EXISTS milestones JSONB;

-- Add check constraint for status if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'smart_cheques_status_check'
    ) THEN
        ALTER TABLE smart_cheques 
        ADD CONSTRAINT smart_cheques_status_check 
        CHECK (status IN ('created', 'locked', 'in_progress', 'completed', 'disputed'));
    END IF;
END
$$;

-- Add check constraint for currency if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'smart_cheques_currency_check'
    ) THEN
        ALTER TABLE smart_cheques 
        ADD CONSTRAINT smart_cheques_currency_check 
        CHECK (currency IN ('USDT', 'USDC', 'e₹'));
    END IF;
END
$$;

-- Create a view for smart cheque analytics
CREATE OR REPLACE VIEW smart_cheque_analytics AS
SELECT 
    status,
    currency,
    COUNT(*) as count,
    SUM(amount) as total_amount,
    AVG(amount) as average_amount,
    MIN(amount) as min_amount,
    MAX(amount) as max_amount,
    MIN(created_at) as first_created,
    MAX(created_at) as last_created
FROM smart_cheques
GROUP BY status, currency;

-- Create a function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_smart_cheque_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger to automatically update updated_at
DROP TRIGGER IF EXISTS update_smart_cheques_updated_at ON smart_cheques;
CREATE TRIGGER update_smart_cheques_updated_at
    BEFORE UPDATE ON smart_cheques
    FOR EACH ROW
    EXECUTE FUNCTION update_smart_cheque_updated_at();