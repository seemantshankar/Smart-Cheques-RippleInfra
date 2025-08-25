CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE wallets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    enterprise_id UUID NOT NULL REFERENCES enterprises(id) ON DELETE CASCADE,
    address VARCHAR(34) NOT NULL UNIQUE,
    public_key TEXT NOT NULL,
    encrypted_private_key TEXT NOT NULL,
    encrypted_seed TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'active', 'suspended', 'deactivated')),
    is_whitelisted BOOLEAN NOT NULL DEFAULT false,
    network_type VARCHAR(10) NOT NULL DEFAULT 'testnet' CHECK (network_type IN ('testnet', 'mainnet')),
    metadata JSONB DEFAULT '{}',
    last_activity TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX idx_wallets_enterprise_id ON wallets(enterprise_id);
CREATE INDEX idx_wallets_address ON wallets(address);
CREATE INDEX idx_wallets_status ON wallets(status);
CREATE INDEX idx_wallets_is_whitelisted ON wallets(is_whitelisted);
CREATE INDEX idx_wallets_network_type ON wallets(network_type);

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_wallets_updated_at 
    BEFORE UPDATE ON wallets 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Add constraint to ensure one active wallet per enterprise per network
CREATE UNIQUE INDEX idx_wallets_enterprise_network_active 
    ON wallets(enterprise_id, network_type) 
    WHERE status = 'active';