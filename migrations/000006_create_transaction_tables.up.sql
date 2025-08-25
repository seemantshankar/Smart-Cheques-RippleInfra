-- Migration: Create transaction management tables
-- Description: Add tables for transaction queue management and batching system

-- Create transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    priority INTEGER NOT NULL DEFAULT 2,
    batch_id VARCHAR(255),
    
    -- Transaction Details
    from_address VARCHAR(255) NOT NULL,
    to_address VARCHAR(255) NOT NULL,
    amount VARCHAR(255) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'XRP',
    fee VARCHAR(255),
    
    -- XRPL Specific Fields
    sequence INTEGER,
    ledger_index INTEGER,
    transaction_hash VARCHAR(255),
    
    -- Escrow Specific Fields
    condition TEXT,
    fulfillment TEXT,
    cancel_after INTEGER,
    finish_after INTEGER,
    offer_sequence INTEGER,
    
    -- Business Context
    smart_cheque_id VARCHAR(255),
    milestone_id VARCHAR(255),
    enterprise_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    
    -- Error Handling
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 3,
    last_error TEXT,
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    scheduled_at TIMESTAMP WITH TIME ZONE,
    processed_at TIMESTAMP WITH TIME ZONE,
    confirmed_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE
);

-- Create transaction_batches table
CREATE TABLE IF NOT EXISTS transaction_batches (
    id VARCHAR(255) PRIMARY KEY,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    priority INTEGER NOT NULL DEFAULT 2,
    
    -- Batch Configuration
    max_transactions INTEGER NOT NULL DEFAULT 10,
    total_fee VARCHAR(255),
    optimized_fee VARCHAR(255),
    fee_savings VARCHAR(255),
    
    -- Processing Details
    transaction_count INTEGER NOT NULL DEFAULT 0,
    success_count INTEGER NOT NULL DEFAULT 0,
    failure_count INTEGER NOT NULL DEFAULT 0,
    
    -- Error Handling
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 3,
    last_error TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);
CREATE INDEX IF NOT EXISTS idx_transactions_batch_id ON transactions(batch_id);
CREATE INDEX IF NOT EXISTS idx_transactions_enterprise_id ON transactions(enterprise_id);
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_type ON transactions(type);
CREATE INDEX IF NOT EXISTS idx_transactions_priority_created ON transactions(priority DESC, created_at ASC);
CREATE INDEX IF NOT EXISTS idx_transactions_smart_cheque_id ON transactions(smart_cheque_id);
CREATE INDEX IF NOT EXISTS idx_transactions_milestone_id ON transactions(milestone_id);
CREATE INDEX IF NOT EXISTS idx_transactions_transaction_hash ON transactions(transaction_hash);
CREATE INDEX IF NOT EXISTS idx_transactions_expires_at ON transactions(expires_at);

CREATE INDEX IF NOT EXISTS idx_transaction_batches_status ON transaction_batches(status);
CREATE INDEX IF NOT EXISTS idx_transaction_batches_priority_created ON transaction_batches(priority DESC, created_at ASC);
CREATE INDEX IF NOT EXISTS idx_transaction_batches_created_at ON transaction_batches(created_at);

-- Add foreign key constraint for batch relationship
ALTER TABLE transactions 
ADD CONSTRAINT fk_transactions_batch_id 
FOREIGN KEY (batch_id) REFERENCES transaction_batches(id) 
ON DELETE SET NULL;

-- Create function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for automatic timestamp updates
CREATE TRIGGER update_transactions_updated_at 
    BEFORE UPDATE ON transactions 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_transaction_batches_updated_at 
    BEFORE UPDATE ON transaction_batches 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Create view for transaction statistics
CREATE OR REPLACE VIEW transaction_stats_view AS
SELECT 
    COUNT(*) as total_transactions,
    COUNT(CASE WHEN status IN ('pending', 'queued', 'batching') THEN 1 END) as pending_transactions,
    COUNT(CASE WHEN status IN ('processing', 'submitted') THEN 1 END) as processing_transactions,
    COUNT(CASE WHEN status = 'confirmed' THEN 1 END) as completed_transactions,
    COUNT(CASE WHEN status IN ('failed', 'cancelled', 'expired') THEN 1 END) as failed_transactions,
    AVG(EXTRACT(EPOCH FROM (confirmed_at - created_at))) as avg_processing_time_seconds,
    MAX(confirmed_at) as last_processed_at
FROM transactions;

-- Create view for batch statistics
CREATE OR REPLACE VIEW batch_stats_view AS
SELECT 
    COUNT(*) as total_batches,
    COUNT(CASE WHEN status IN ('pending', 'batching') THEN 1 END) as pending_batches,
    COUNT(CASE WHEN status = 'processing' THEN 1 END) as processing_batches,
    COUNT(CASE WHEN status = 'confirmed' THEN 1 END) as completed_batches,
    COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_batches,
    AVG(transaction_count) as avg_batch_size,
    AVG(EXTRACT(EPOCH FROM (completed_at - created_at))) as avg_batch_processing_time_seconds
FROM transaction_batches;

-- Create view for hourly transaction statistics
CREATE OR REPLACE VIEW hourly_transaction_stats AS
SELECT 
    date_trunc('hour', created_at) as hour,
    COUNT(*) as transaction_count,
    COUNT(CASE WHEN status = 'confirmed' THEN 1 END) as success_count,
    COUNT(CASE WHEN status IN ('failed', 'cancelled', 'expired') THEN 1 END) as failure_count,
    AVG(CASE WHEN confirmed_at IS NOT NULL THEN EXTRACT(EPOCH FROM (confirmed_at - created_at)) END) as avg_processing_time,
    SUM(CASE WHEN fee IS NOT NULL AND fee != '' THEN CAST(fee AS BIGINT) ELSE 0 END) as total_fees
FROM transactions 
WHERE created_at >= NOW() - INTERVAL '24 hours'
GROUP BY date_trunc('hour', created_at)
ORDER BY hour DESC;

-- Add comments for documentation
COMMENT ON TABLE transactions IS 'Stores all blockchain transactions in the queue management system';
COMMENT ON TABLE transaction_batches IS 'Stores transaction batches for fee optimization and bulk processing';

COMMENT ON COLUMN transactions.status IS 'Current status: pending, queued, batching, batched, processing, submitted, confirmed, failed, cancelled, expired';
COMMENT ON COLUMN transactions.priority IS 'Priority level: 1=low, 2=normal, 3=high, 4=critical';
COMMENT ON COLUMN transactions.type IS 'Transaction type: escrow_create, escrow_finish, escrow_cancel, payment, wallet_setup';
COMMENT ON COLUMN transactions.amount IS 'Transaction amount in smallest unit (drops for XRP)';
COMMENT ON COLUMN transactions.fee IS 'Transaction fee in drops';
COMMENT ON COLUMN transactions.metadata IS 'Additional transaction-specific data stored as JSON';

COMMENT ON COLUMN transaction_batches.status IS 'Batch status: pending, batching, batched, processing, confirmed, failed';
COMMENT ON COLUMN transaction_batches.total_fee IS 'Sum of individual transaction fees before optimization';
COMMENT ON COLUMN transaction_batches.optimized_fee IS 'Actual total fee after batch optimization';
COMMENT ON COLUMN transaction_batches.fee_savings IS 'Amount saved through batch optimization';

-- Create sample data for testing (only in development)
-- INSERT INTO transaction_batches (id, status, priority, max_transactions) 
-- VALUES ('sample-batch-001', 'pending', 2, 10);

-- INSERT INTO transactions (id, type, status, priority, from_address, to_address, amount, currency, enterprise_id, user_id)
-- VALUES (
--     'sample-tx-001',
--     'payment',
--     'pending',
--     2,
--     'rSampleSender123',
--     'rSampleReceiver456',
--     '1000000',
--     'XRP',
--     'sample-enterprise-001',
--     'sample-user-001'
-- );