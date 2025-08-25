#!/bin/bash

# Cloud PostgreSQL Setup Script for Smart Payment Infrastructure
# This script helps you configure a cloud PostgreSQL database

echo "🚀 Smart Payment Infrastructure - Cloud Database Setup"
echo "======================================================="
echo ""

# Function to update .env file
update_env() {
    local key=$1
    local value=$2
    local env_file=".env"
    
    if grep -q "^${key}=" "$env_file"; then
        # Key exists, update it
        if [[ "$OSTYPE" == "darwin"* ]]; then
            # macOS
            sed -i '' "s|^${key}=.*|${key}=${value}|" "$env_file"
        else
            # Linux
            sed -i "s|^${key}=.*|${key}=${value}|" "$env_file"
        fi
    else
        # Key doesn't exist, add it
        echo "${key}=${value}" >> "$env_file"
    fi
    echo "✅ Updated ${key} in .env file"
}

# Function to test database connection
test_connection() {
    echo ""
    echo "🧪 Testing database connection..."
    if go run cmd/db-migrate/main.go -action=version > /dev/null 2>&1; then
        echo "✅ Database connection successful!"
        return 0
    else
        echo "❌ Database connection failed. Please check your connection string."
        return 1
    fi
}

# Function to run migrations
run_migrations() {
    echo ""
    echo "🔄 Running database migrations..."
    if go run cmd/db-migrate/main.go -action=up; then
        echo "✅ Database migrations completed successfully!"
        return 0
    else
        echo "❌ Database migrations failed."
        return 1
    fi
}

# Function to seed development data
seed_data() {
    echo ""
    echo "🌱 Seeding development data..."
    if go run cmd/db-migrate/main.go -action=seed; then
        echo "✅ Development data seeded successfully!"
        return 0
    else
        echo "❌ Failed to seed development data."
        return 1
    fi
}

# Main setup process
echo "Choose your cloud PostgreSQL provider:"
echo "1) Supabase (🚀 RECOMMENDED FOR PMF - Free 500MB forever, 2-min setup)"
echo "2) Neon (Free 3GB forever, simple setup)"
echo "3) Amazon Aurora DSQL (NEW - True serverless, 100k DPUs + 1GB free)"
echo "4) AWS RDS (Free 20GB for 12 months)"
echo "5) Azure Database for PostgreSQL (Free 32GB for 12 months)"
echo "6) Google Cloud Platform (Self-hosted, 30GB forever free)"
echo "7) Railway (Free $5 credit)"
echo "8) Custom connection string"
echo ""
echo "💡 For achieving Product-Market Fit (PMF): Choose Supabase (#1)"
echo ""

read -p "Enter your choice (1-8): " choice

case $choice in
    1)
        echo ""
        echo "🚀 Supabase Setup Instructions (Perfect for PMF!):"
        echo "1. Go to https://supabase.com"
        echo "2. Click 'Start your project' and sign up with GitHub (no credit card!)"
        echo "3. Create project: 'Smart Payment Infrastructure'"
        echo "4. Generate a strong database password (save it!)"
        echo "5. Wait ~2 minutes for project creation"
        echo "6. Go to Settings → Database → Connection string (URI format)"
        echo "7. Copy the string and replace [YOUR-PASSWORD] with your password"
        echo ""
        echo "✨ Bonus: Real-time features, built-in auth, visual dashboard included!"
        echo "📖 Detailed guide: docs/supabase-setup.md"
        echo ""
        ;;
    2)
        echo ""
        echo "📋 Neon Setup Instructions:"
        echo "1. Go to https://neon.tech"
        echo "2. Sign up with GitHub/Google"
        echo "3. Create a new project named 'smart-payment-infrastructure'"
        echo "4. Copy the connection string from the dashboard"
        echo ""
        ;;
    3)
        echo ""
        echo "📋 Amazon Aurora DSQL Setup Instructions:"
        echo "1. Go to https://aws.amazon.com and create free account"
        echo "2. Navigate to Aurora DSQL service"
        echo "3. Create cluster with PostgreSQL compatibility"
        echo "4. Create database named 'smart_payment'"
        echo "5. Configure IAM authentication or password auth"
        echo "6. Use format: postgresql://username:password@cluster.dsql.region.on.aws:5432/smart_payment?sslmode=require"
        echo ""
        echo "🆕 Aurora DSQL: True serverless, 100k DPUs + 1GB free monthly, unlimited scale"
        echo ""
        ;;
    4)
        echo ""
        echo "📋 AWS RDS Setup Instructions:"
        echo "1. Go to https://aws.amazon.com and create free account"
        echo "2. Navigate to RDS service"
        echo "3. Create PostgreSQL database with 'Free tier' template"
        echo "4. Use db.t3.micro instance (free tier eligible)"
        echo "5. Set 20GB storage (maximum for free tier)"
        echo "6. Enable public access and configure security group"
        echo "7. Use format: postgresql://postgres:password@endpoint.region.rds.amazonaws.com:5432/smart_payment?sslmode=require"
        echo ""
        echo "⚠️  AWS Free Tier: 750 hours/month for 12 months, then ~$15/month"
        echo ""
        ;;
    5)
        echo ""
        echo "📋 Azure Database for PostgreSQL Setup Instructions:"
        echo "1. Go to https://azure.microsoft.com/free and create account"
        echo "2. Navigate to 'Azure Database for PostgreSQL flexible servers'"
        echo "3. Create database with Burstable B1ms instance"
        echo "4. Set 32GB storage (maximum for free tier)"
        echo "5. Enable public access and configure networking"
        echo "6. Use format: postgresql://postgres:password@server-name.postgres.database.azure.com:5432/smart_payment?sslmode=require"
        echo ""
        echo "⚠️  Azure Free Tier: 750 hours/month for 12 months, then ~$25-35/month"
        echo ""
        ;;
    6)
        echo ""
        echo "📋 Google Cloud Platform (Self-hosted) Setup Instructions:"
        echo "1. Go to https://cloud.google.com and create account"
        echo "2. Create e2-micro VM instance in us-west1, us-central1, or us-east1"
        echo "3. Install PostgreSQL on the VM (see docs/database-setup.md for commands)"
        echo "4. Configure firewall to allow port 5432"
        echo "5. Use format: postgresql://appuser:password@EXTERNAL_IP:5432/smart_payment?sslmode=prefer"
        echo ""
        echo "✅ GCP Always Free: 30GB, forever free, but requires manual setup"
        echo ""
        ;;
    7)
        echo ""
        echo "📋 Railway Setup Instructions:"
        echo "1. Go to https://railway.app"
        echo "2. Sign up with GitHub"
        echo "3. Create new project → Add PostgreSQL"
        echo "4. Copy the Postgres Connection URL"
        echo ""
        ;;
    8)
        echo ""
        echo "📋 Custom Setup:"
        echo "Enter your PostgreSQL connection string"
        echo ""
        ;;
    *)
        echo "❌ Invalid choice. Exiting."
        exit 1
        ;;
esac

# Get connection string from user
echo "📝 Enter your PostgreSQL connection string:"
echo "Example: postgresql://username:password@host:port/database?sslmode=require"
echo ""
read -p "Connection string: " connection_string

if [ -z "$connection_string" ]; then
    echo "❌ Connection string cannot be empty. Exiting."
    exit 1
fi

# Update .env file
echo ""
echo "📝 Updating configuration..."
update_env "POSTGRES_URL" "$connection_string"

# Test the connection
if test_connection; then
    echo ""
    read -p "🚀 Would you like to run database migrations? (y/n): " run_migrations_choice
    
    if [[ $run_migrations_choice =~ ^[Yy]$ ]]; then
        if run_migrations; then
            echo ""
            read -p "🌱 Would you like to seed development data? (y/n): " seed_choice
            
            if [[ $seed_choice =~ ^[Yy]$ ]]; then
                seed_data
            fi
        fi
    fi
    
    echo ""
    echo "🎉 Setup completed successfully!"
    echo ""
    echo "Next steps:"
    echo "1. Run integration tests: go test ./test/integration -v"
    echo "2. Start the identity service: go run cmd/identity-service/main.go"
    echo "3. Test the API endpoints with your favorite HTTP client"
    echo ""
    
else
    echo ""
    echo "❌ Setup failed. Please check your connection string and try again."
    echo ""
    echo "Common issues:"
    echo "- Incorrect username/password"
    echo "- Wrong hostname or port"
    echo "- Database doesn't exist"
    echo "- SSL mode requirements (try adding ?sslmode=require)"
    echo ""
fi