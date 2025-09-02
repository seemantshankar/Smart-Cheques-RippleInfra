# 🚀 **Supabase Setup Guide for Smart-Cheques Platform**

## 📋 **Overview**
This guide will help you set up Supabase as your PostgreSQL database for the Smart-Cheques platform. Supabase provides a generous free tier perfect for development and PMF validation.

## ⚡ **Quick Setup (5 minutes)**

### Step 1: Create Supabase Account
1. Go to [https://supabase.com](https://supabase.com)
2. Click **"Start your project"**
3. Sign up with GitHub (recommended) or email
4. **No credit card required!** 🎉

### Step 2: Create New Project
1. Click **"New project"**
2. **Organization**: Personal (default)
3. **Project name**: `Smart Payment Infrastructure`
4. **Database password**: Generate strong password (save this!)
5. **Region**: Choose closest to you (e.g., US East, Europe)
6. Click **"Create new project"**
7. ⏱️ Wait ~2 minutes for setup

### Step 3: Get Your Credentials

#### 📍 **Get Database Connection String**
1. Go to **Settings** → **Database**
2. Scroll down to **"Connection string"**
3. Copy the **URI** format connection string
4. Replace `[YOUR-PASSWORD]` with your actual database password

**Example connection string:**
```
postgresql://postgres:your-actual-password@db.abc123def456.supabase.co:5432/postgres?sslmode=require
```

#### 🔑 **Get API Keys**
1. Go to **Settings** → **API**
2. Copy the following keys:
   - **Project URL**: `https://abc123def456.supabase.co`
   - **anon public key**: `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...`
   - **service_role secret key**: `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...`

#### 📝 **Get Project Reference**
1. Your project reference is in the URL: `https://abc123def456.supabase.co`
2. The reference is: `abc123def456`

## 🔧 **Update Your .env File**

Replace the placeholder values in your `.env` file with your actual Supabase credentials:

```bash
# =============================================================================
# SUPABASE CONFIGURATION
# =============================================================================

# Database Connection (from Settings → Database)
POSTGRES_URL=postgresql://postgres:your-actual-password@db.abc123def456.supabase.co:5432/postgres?sslmode=require
DATABASE_URL=postgresql://postgres:your-actual-password@db.abc123def456.supabase.co:5432/postgres?sslmode=require

# Project Details (from URL and Settings → API)
SUPABASE_PROJECT_REF=abc123def456
SUPABASE_URL=https://abc123def456.supabase.co

# API Keys (from Settings → API)
SUPABASE_ANON_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
SUPABASE_SERVICE_ROLE_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

# Database Password (separate from API keys)
SUPABASE_DB_PASSWORD=your-actual-database-password
```

## 🧪 **Test Your Connection**

### Test Database Connection
```bash
# Test the connection
go run cmd/db-migrate/main.go -action=version

# Expected output: Migration version information
```

### Run Migrations
```bash
# Run all database migrations
go run cmd/db-migrate/main.go -action=up

# Expected output: Migration success messages
```

### Seed Development Data
```bash
# Add sample data for development
go run cmd/db-migrate/main.go -action=seed
```

## 🎯 **Verify Setup**

### Check Database Tables
```bash
# List all tables
go run cmd/list-tables/main.go
```

### Test Full Stack
```bash
# Build all services
make build

# Start all services
docker-compose up -d

# Test health endpoints
curl http://localhost:8000/health
curl http://localhost:8001/health
curl http://localhost:8002/health
curl http://localhost:8003/health
curl http://localhost:8004/health
```

## 📊 **Monitor Usage (Stay in Free Tier)**

### Free Tier Limits
- **💾 Database**: 500MB PostgreSQL database
- **🔗 API Requests**: 50,000 API requests/month
- **🔐 Auth Users**: 50,000 monthly active users
- **📁 Storage**: 1GB file storage
- **🌐 Bandwidth**: 2GB bandwidth/month

### Check Usage
1. Go to **Settings** → **Usage** in Supabase dashboard
2. Monitor your usage against limits
3. Set up email alerts at 80% usage

## 🔧 **Troubleshooting**

### "Connection refused"
```bash
# Check your connection string format
echo $POSTGRES_URL

# Test with a simple connection
psql "$POSTGRES_URL" -c "SELECT version();"
```

### "Authentication failed"
- ✅ Double-check your database password
- ✅ Ensure you're using `postgres` as the username
- ✅ Try resetting the database password in Supabase dashboard

### "Database does not exist"
- ✅ Use `postgres` as the database name (this is correct for Supabase)
- ✅ Don't create custom database names

### Migrations Fail
```bash
# Check existing tables in Supabase dashboard
# Drop conflicting tables manually if needed

# Run migrations one by one
go run cmd/db-migrate/main.go -action=up -limit=1
```

## 🎉 **Success Indicators**

When everything is working correctly:
- ✅ `go run cmd/db-migrate/main.go -action=version` returns version info
- ✅ `go run cmd/db-migrate/main.go -action=up` completes without errors
- ✅ All services start without database connection errors
- ✅ Health endpoints return `200 OK`

## 🚀 **Next Steps**

1. ✅ **Database Connected**: Your Supabase database is ready
2. ✅ **Migrations Run**: Schema is set up
3. ✅ **Services Configured**: All microservices can connect
4. 🔄 **Start Development**: Begin building payment features
5. 📈 **Monitor Growth**: Track usage and plan for scaling

## 💡 **Pro Tips**

- **Use Supabase Dashboard**: Visual table editor for rapid development
- **Enable Real-time**: Perfect for payment status updates
- **Set up Row Level Security**: Secure sensitive payment data
- **Use Edge Functions**: For webhook handlers and payment processing

---

**🎯 Ready for PMF development!** Your Smart-Cheques platform now has a robust, scalable database foundation.
