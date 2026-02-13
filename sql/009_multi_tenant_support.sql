-- Create companies table
CREATE TABLE companies (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    address TEXT,
    phone_number TEXT,
    email TEXT UNIQUE,
    status TEXT DEFAULT 'pending', -- pending, active, suspended
    commission_rate DECIMAL DEFAULT 10.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add company_id to drivers
ALTER TABLE drivers ADD COLUMN company_id UUID REFERENCES companies(id);

-- Create admin_users table for multi-tenant management
CREATE TABLE admin_users (
    id UUID PRIMARY KEY,
    company_id UUID REFERENCES companies(id), -- NULL for Super Admin (Platform Owner)
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL, -- 'super_admin', 'company_admin'
    name TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Register initial Super Admin (Dummy for reference)
-- Note: In a real app, passwords should be hashed.
INSERT INTO admin_users (id, username, password_hash, role, name) 
VALUES (gen_random_uuid(), 'admin', 'admin123', 'super_admin', 'Platform Owner');
