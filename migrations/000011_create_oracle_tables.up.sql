-- Create oracle providers table
CREATE TABLE IF NOT EXISTS oracle_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL,
    endpoint TEXT,
    auth_config JSONB,
    rate_limit_config JSONB,
    is_active BOOLEAN DEFAULT true,
    reliability DECIMAL(3,2) DEFAULT 1.00,
    response_time BIGINT DEFAULT 0,
    capabilities TEXT[] DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT chk_oracle_provider_type CHECK (type IN ('api', 'webhook', 'manual')),
    CONSTRAINT chk_oracle_provider_reliability CHECK (reliability >= 0 AND reliability <= 1)
);

-- Create oracle requests table
CREATE TABLE IF NOT EXISTS oracle_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id UUID NOT NULL,
    condition TEXT NOT NULL,
    context_data JSONB,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    result BOOLEAN,
    confidence DECIMAL(3,2),
    evidence BYTEA,
    metadata JSONB,
    verified_at TIMESTAMP WITH TIME ZONE,
    proof_hash VARCHAR(64),
    retry_count INTEGER DEFAULT 0,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    cached_until TIMESTAMP WITH TIME ZONE,

    CONSTRAINT fk_oracle_requests_provider FOREIGN KEY (provider_id)
        REFERENCES oracle_providers(id) ON DELETE CASCADE,
    CONSTRAINT chk_oracle_requests_status CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cached')),
    CONSTRAINT chk_oracle_requests_confidence CHECK (confidence IS NULL OR (confidence >= 0 AND confidence <= 1))
);

-- Add indexes for oracle providers
CREATE INDEX IF NOT EXISTS idx_oracle_providers_type ON oracle_providers(type);
CREATE INDEX IF NOT EXISTS idx_oracle_providers_active ON oracle_providers(is_active);
CREATE INDEX IF NOT EXISTS idx_oracle_providers_reliability ON oracle_providers(reliability);
CREATE INDEX IF NOT EXISTS idx_oracle_providers_created_at ON oracle_providers(created_at);

-- Add indexes for oracle requests
CREATE INDEX IF NOT EXISTS idx_oracle_requests_provider_id ON oracle_requests(provider_id);
CREATE INDEX IF NOT EXISTS idx_oracle_requests_status ON oracle_requests(status);
CREATE INDEX IF NOT EXISTS idx_oracle_requests_condition ON oracle_requests(condition);
CREATE INDEX IF NOT EXISTS idx_oracle_requests_created_at ON oracle_requests(created_at);
CREATE INDEX IF NOT EXISTS idx_oracle_requests_cached_until ON oracle_requests(cached_until);
CREATE INDEX IF NOT EXISTS idx_oracle_requests_verified_at ON oracle_requests(verified_at);

-- Add GIN indexes for JSONB fields
CREATE INDEX IF NOT EXISTS idx_oracle_providers_capabilities_gin ON oracle_providers USING gin(capabilities);
CREATE INDEX IF NOT EXISTS idx_oracle_requests_context_data_gin ON oracle_requests USING gin(context_data);
CREATE INDEX IF NOT EXISTS idx_oracle_requests_metadata_gin ON oracle_requests USING gin(metadata);

-- Create triggers for automatic updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply triggers to oracle tables
DROP TRIGGER IF EXISTS update_oracle_providers_updated_at ON oracle_providers;
CREATE TRIGGER update_oracle_providers_updated_at
    BEFORE UPDATE ON oracle_providers
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_oracle_requests_updated_at ON oracle_requests;
CREATE TRIGGER update_oracle_requests_updated_at
    BEFORE UPDATE ON oracle_requests
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create views for common queries

-- View for oracle provider statistics
CREATE OR REPLACE VIEW oracle_provider_stats AS
SELECT
    op.id,
    op.name,
    op.type,
    op.is_active,
    op.reliability,
    op.response_time,
    COUNT(CASE WHEN orq.status = 'completed' THEN 1 END) as successful_requests,
    COUNT(CASE WHEN orq.status = 'failed' THEN 1 END) as failed_requests,
    COUNT(*) as total_requests,
    CASE
        WHEN COUNT(*) > 0 THEN
            COUNT(CASE WHEN orq.status = 'completed' THEN 1 END) * 100.0 / COUNT(*)
        ELSE 0
    END as success_rate,
    AVG(orq.confidence) as average_confidence
FROM oracle_providers op
LEFT JOIN oracle_requests orq ON op.id = orq.provider_id
GROUP BY op.id, op.name, op.type, op.is_active, op.reliability, op.response_time;

-- View for cached oracle responses
CREATE OR REPLACE VIEW cached_oracle_responses AS
SELECT
    id,
    condition,
    result,
    confidence,
    verified_at,
    cached_until,
    EXTRACT(EPOCH FROM (cached_until - NOW())) as seconds_until_expiration
FROM oracle_requests
WHERE status = 'cached' AND cached_until > NOW();

-- View for oracle request trends
CREATE OR REPLACE VIEW oracle_request_trends AS
SELECT
    DATE(created_at) as request_date,
    status,
    COUNT(*) as request_count,
    AVG(confidence) as average_confidence
FROM oracle_requests
GROUP BY DATE(created_at), status
ORDER BY request_date DESC, status;