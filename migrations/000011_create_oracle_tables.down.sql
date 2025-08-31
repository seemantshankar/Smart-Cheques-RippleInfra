-- Drop views
DROP VIEW IF EXISTS oracle_request_trends;
DROP VIEW IF EXISTS cached_oracle_responses;
DROP VIEW IF EXISTS oracle_provider_stats;

-- Drop triggers
DROP TRIGGER IF EXISTS update_oracle_requests_updated_at ON oracle_requests;
DROP TRIGGER IF EXISTS update_oracle_providers_updated_at ON oracle_providers;

-- Drop tables
DROP TABLE IF EXISTS oracle_requests;
DROP TABLE IF EXISTS oracle_providers;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();