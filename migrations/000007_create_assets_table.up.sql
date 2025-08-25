-- Migration: Create assets and balances tables
-- Description: Add tables for asset management including supported currencies and balance tracking

-- Create supported_assets table
CREATE TABLE IF NOT EXISTS supported_assets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    currency_code VARCHAR(10) NOT NULL UNIQUE,
    currency_name VARCHAR(100) NOT NULL,
    asset_type VARCHAR(20) NOT NULL,
    issuer_address VARCHAR(100),
    currency_hex VARCHAR(40),
    decimal_places INTEGER NOT NULL DEFAULT 8,
    minimum_amount VARCHAR(255) NOT NULL DEFAULT '0',
    maximum_amount VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT true,
    
    -- XRPL specific fields
    trust_line_limit VARCHAR(255),
    transfer_fee DECIMAL(8,6) DEFAULT 0.000000,
    global_freeze BOOLEAN DEFAULT false,
    no_freeze BOOLEAN DEFAULT false,
    
    -- Metadata
    description TEXT,
    icon_url VARCHAR(500),
    documentation_url VARCHAR(500),
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT asset_type_check CHECK (asset_type IN ('native', 'stablecoin', 'cbdc', 'wrapped')),
    CONSTRAINT decimal_places_check CHECK (decimal_places >= 0 AND decimal_places <= 18)
);

-- Create enterprise_balances table
CREATE TABLE IF NOT EXISTS enterprise_balances (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    enterprise_id UUID NOT NULL REFERENCES enterprises(id) ON DELETE CASCADE,
    currency_code VARCHAR(10) NOT NULL,
    
    -- Balance tracking
    available_balance VARCHAR(255) NOT NULL DEFAULT '0',
    reserved_balance VARCHAR(255) NOT NULL DEFAULT '0',
    total_balance VARCHAR(255) NOT NULL DEFAULT '0',
    
    -- XRPL tracking
    xrpl_balance VARCHAR(255) NOT NULL DEFAULT '0',
    last_xrpl_sync TIMESTAMP WITH TIME ZONE,
    
    -- Limits and controls
    daily_limit VARCHAR(255),
    monthly_limit VARCHAR(255),
    max_transaction_amount VARCHAR(255),
    
    -- Status and metadata
    is_frozen BOOLEAN NOT NULL DEFAULT false,
    freeze_reason TEXT,
    last_transaction_at TIMESTAMP WITH TIME ZONE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    UNIQUE(enterprise_id, currency_code),
    FOREIGN KEY (currency_code) REFERENCES supported_assets(currency_code)
);

-- Create asset_transactions table for internal tracking
CREATE TABLE IF NOT EXISTS asset_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    enterprise_id UUID NOT NULL REFERENCES enterprises(id),
    currency_code VARCHAR(10) NOT NULL,
    transaction_type VARCHAR(20) NOT NULL,
    
    -- Transaction details
    amount VARCHAR(255) NOT NULL,
    fee VARCHAR(255) DEFAULT '0',
    reference_id VARCHAR(255),
    external_tx_hash VARCHAR(255),
    
    -- Balances before/after
    balance_before VARCHAR(255) NOT NULL,
    balance_after VARCHAR(255) NOT NULL,
    
    -- Status and metadata
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    description TEXT,
    metadata JSONB DEFAULT '{}',
    
    -- Approval workflow
    approved_by UUID,
    approved_at TIMESTAMP WITH TIME ZONE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT transaction_type_check CHECK (transaction_type IN ('deposit', 'withdrawal', 'transfer_in', 'transfer_out', 'escrow_lock', 'escrow_release', 'fee', 'adjustment')),
    CONSTRAINT status_check CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled')),
    FOREIGN KEY (currency_code) REFERENCES supported_assets(currency_code)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_supported_assets_currency_code ON supported_assets(currency_code);
CREATE INDEX IF NOT EXISTS idx_supported_assets_is_active ON supported_assets(is_active);
CREATE INDEX IF NOT EXISTS idx_supported_assets_asset_type ON supported_assets(asset_type);

CREATE INDEX IF NOT EXISTS idx_enterprise_balances_enterprise_id ON enterprise_balances(enterprise_id);
CREATE INDEX IF NOT EXISTS idx_enterprise_balances_currency_code ON enterprise_balances(currency_code);
CREATE INDEX IF NOT EXISTS idx_enterprise_balances_enterprise_currency ON enterprise_balances(enterprise_id, currency_code);

CREATE INDEX IF NOT EXISTS idx_asset_transactions_enterprise_id ON asset_transactions(enterprise_id);
CREATE INDEX IF NOT EXISTS idx_asset_transactions_currency_code ON asset_transactions(currency_code);
CREATE INDEX IF NOT EXISTS idx_asset_transactions_type ON asset_transactions(transaction_type);
CREATE INDEX IF NOT EXISTS idx_asset_transactions_status ON asset_transactions(status);
CREATE INDEX IF NOT EXISTS idx_asset_transactions_created_at ON asset_transactions(created_at);
CREATE INDEX IF NOT EXISTS idx_asset_transactions_reference_id ON asset_transactions(reference_id);

-- Create triggers for automatic timestamp updates
CREATE TRIGGER update_supported_assets_updated_at 
    BEFORE UPDATE ON supported_assets 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_enterprise_balances_updated_at 
    BEFORE UPDATE ON enterprise_balances 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_asset_transactions_updated_at 
    BEFORE UPDATE ON asset_transactions 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Create views for balance reporting
CREATE OR REPLACE VIEW enterprise_balance_summary AS
SELECT 
    eb.enterprise_id,
    e.legal_name as enterprise_name,
    eb.currency_code,
    sa.currency_name,
    eb.available_balance,
    eb.reserved_balance,
    eb.total_balance,
    eb.xrpl_balance,
    eb.is_frozen,
    eb.last_transaction_at,
    eb.last_xrpl_sync
FROM enterprise_balances eb
JOIN enterprises e ON eb.enterprise_id = e.id
JOIN supported_assets sa ON eb.currency_code = sa.currency_code
WHERE sa.is_active = true;

-- Insert default supported assets
INSERT INTO supported_assets (currency_code, currency_name, asset_type, decimal_places, minimum_amount, description) VALUES
('XRP', 'XRP', 'native', 6, '1', 'Native XRP cryptocurrency'),
('USDT', 'Tether USD', 'stablecoin', 6, '0.01', 'USD-pegged stablecoin issued by Tether'),
('USDC', 'USD Coin', 'stablecoin', 6, '0.01', 'USD-pegged stablecoin issued by Centre'),
('e₹', 'Digital Rupee', 'cbdc', 2, '0.01', 'Central Bank Digital Currency issued by Reserve Bank of India')
ON CONFLICT (currency_code) DO NOTHING;

-- Add table comments
COMMENT ON TABLE supported_assets IS 'Registry of all supported assets and currencies in the platform';
COMMENT ON TABLE enterprise_balances IS 'Tracks balance information for each enterprise by currency';
COMMENT ON TABLE asset_transactions IS 'Internal ledger of all asset movements and transactions';

COMMENT ON COLUMN supported_assets.currency_hex IS 'Hexadecimal representation of currency code for XRPL (for issued currencies)';
COMMENT ON COLUMN supported_assets.issuer_address IS 'XRPL address of the asset issuer (null for XRP)';
COMMENT ON COLUMN enterprise_balances.available_balance IS 'Balance available for transactions';
COMMENT ON COLUMN enterprise_balances.reserved_balance IS 'Balance reserved in escrows or pending transactions';
COMMENT ON COLUMN enterprise_balances.total_balance IS 'Total balance (available + reserved)';
COMMENT ON COLUMN asset_transactions.reference_id IS 'Reference to related entity (smart_cheque_id, milestone_id, etc.)';