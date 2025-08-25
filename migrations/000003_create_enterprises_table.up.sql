-- Create enterprises table
CREATE TABLE enterprises (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    legal_name VARCHAR(255) NOT NULL,
    trade_name VARCHAR(255),
    registration_number VARCHAR(100) UNIQUE NOT NULL,
    tax_id VARCHAR(100) NOT NULL,
    jurisdiction VARCHAR(100) NOT NULL,
    business_type VARCHAR(100) NOT NULL,
    industry VARCHAR(100) NOT NULL,
    website VARCHAR(255),
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(50) NOT NULL,
    
    -- Address fields
    street1 VARCHAR(255) NOT NULL,
    street2 VARCHAR(255),
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100) NOT NULL,
    postal_code VARCHAR(20) NOT NULL,
    country VARCHAR(100) NOT NULL,
    
    -- Status fields
    kyb_status VARCHAR(50) DEFAULT 'pending' CHECK (kyb_status IN ('pending', 'in_review', 'verified', 'rejected', 'suspended')),
    compliance_status VARCHAR(50) DEFAULT 'pending' CHECK (compliance_status IN ('pending', 'compliant', 'non_compliant', 'under_review')),
    
    -- XRPL integration
    xrpl_wallet VARCHAR(100),
    
    -- Compliance fields
    aml_risk_score INTEGER DEFAULT 0,
    sanctions_check_date TIMESTAMP WITH TIME ZONE,
    pep_check_date TIMESTAMP WITH TIME ZONE,
    compliance_officer VARCHAR(255),
    last_review_date TIMESTAMP WITH TIME ZONE,
    next_review_date TIMESTAMP WITH TIME ZONE,
    
    -- Metadata
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    verified_at TIMESTAMP WITH TIME ZONE
);

-- Create authorized_representatives table
CREATE TABLE authorized_representatives (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    enterprise_id UUID NOT NULL REFERENCES enterprises(id) ON DELETE CASCADE,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(50) NOT NULL,
    position VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create enterprise_documents table
CREATE TABLE enterprise_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    enterprise_id UUID NOT NULL REFERENCES enterprises(id) ON DELETE CASCADE,
    document_type VARCHAR(100) NOT NULL CHECK (document_type IN (
        'certificate_of_incorporation',
        'business_license',
        'tax_certificate',
        'bank_statement',
        'director_id',
        'proof_of_address',
        'articles_of_association',
        'memorandum_of_association'
    )),
    file_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'verified', 'rejected')),
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    verified_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for better performance
CREATE INDEX idx_enterprises_registration_number ON enterprises(registration_number);
CREATE INDEX idx_enterprises_kyb_status ON enterprises(kyb_status);
CREATE INDEX idx_enterprises_compliance_status ON enterprises(compliance_status);
CREATE INDEX idx_enterprises_jurisdiction ON enterprises(jurisdiction);
CREATE INDEX idx_enterprises_created_at ON enterprises(created_at);

CREATE INDEX idx_authorized_representatives_enterprise_id ON authorized_representatives(enterprise_id);
CREATE INDEX idx_authorized_representatives_email ON authorized_representatives(email);

CREATE INDEX idx_enterprise_documents_enterprise_id ON enterprise_documents(enterprise_id);
CREATE INDEX idx_enterprise_documents_document_type ON enterprise_documents(document_type);
CREATE INDEX idx_enterprise_documents_status ON enterprise_documents(status);

-- Create updated_at trigger for enterprises table
CREATE TRIGGER update_enterprises_updated_at BEFORE UPDATE ON enterprises
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();