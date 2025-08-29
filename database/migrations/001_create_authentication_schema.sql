-- ============================================================================
-- Database Migration Script for Authentication System
-- Version: 1.0.0
-- Created: August 23, 2025
-- Description: Initial migration to create authentication schema and tables
-- ============================================================================

-- Migration: 001_create_authentication_schema
-- Up Migration
BEGIN;

-- Create authentication schema
CREATE SCHEMA IF NOT EXISTS authentication;

-- Set default schema for current session
SET search_path TO authentication, public;

-- Create extension for UUID generation if not exists
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- Create Tables
-- ============================================================================

-- Users table
CREATE TABLE authentication.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    email_verified_at TIMESTAMP WITH TIME ZONE NULL,
    password_hash VARCHAR(255) NOT NULL,
    phone VARCHAR(20) UNIQUE NULL,
    phone_verified_at TIMESTAMP WITH TIME ZONE NULL,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended', 'deleted')),
    last_login_at TIMESTAMP WITH TIME ZONE NULL,
    last_login_ip INET NULL,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE NULL,
    two_factor_enabled BOOLEAN DEFAULT FALSE,
    two_factor_secret VARCHAR(255) NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

-- Profiles table
CREATE TABLE authentication.profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES authentication.users(id) ON DELETE CASCADE,
    first_name VARCHAR(100) NULL,
    last_name VARCHAR(100) NULL,
    display_name VARCHAR(200) NULL,
    avatar_url VARCHAR(500) NULL,
    bio TEXT NULL,
    date_of_birth DATE NULL,
    gender VARCHAR(10) CHECK (gender IN ('male', 'female', 'other', 'prefer_not_to_say')),
    country VARCHAR(2) NULL,
    state VARCHAR(100) NULL,
    city VARCHAR(100) NULL,
    address TEXT NULL,
    postal_code VARCHAR(20) NULL,
    timezone VARCHAR(50) DEFAULT 'UTC',
    language VARCHAR(5) DEFAULT 'en',
    website_url VARCHAR(500) NULL,
    social_links JSONB NULL,
    preferences JSONB NULL,
    metadata JSONB NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Token access table
CREATE TABLE authentication.token_access (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES authentication.users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    token_type VARCHAR(20) NOT NULL CHECK (token_type IN ('access', 'refresh', 'api_key', 'verification')),
    token_name VARCHAR(100) NULL,
    scope TEXT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    last_used_at TIMESTAMP WITH TIME ZONE NULL,
    last_used_ip INET NULL,
    user_agent TEXT NULL,
    device_info JSONB NULL,
    revoked BOOLEAN DEFAULT FALSE,
    revoked_at TIMESTAMP WITH TIME ZONE NULL,
    revoked_by UUID NULL REFERENCES authentication.users(id),
    revoke_reason VARCHAR(255) NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Password resets table
CREATE TABLE authentication.password_resets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES authentication.users(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    used_at TIMESTAMP WITH TIME ZONE NULL,
    used_ip INET NULL,
    user_agent TEXT NULL,
    attempts INTEGER DEFAULT 0,
    max_attempts INTEGER DEFAULT 3,
    blocked_until TIMESTAMP WITH TIME ZONE NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Permissions table
CREATE TABLE authentication.permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES authentication.users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL,
    permission VARCHAR(100) NOT NULL,
    resource VARCHAR(100) NULL,
    resource_id UUID NULL,
    granted_by UUID NULL REFERENCES authentication.users(id),
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NULL,
    revoked BOOLEAN DEFAULT FALSE,
    revoked_at TIMESTAMP WITH TIME ZONE NULL,
    revoked_by UUID NULL REFERENCES authentication.users(id),
    conditions JSONB NULL,
    metadata JSONB NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- Create Indexes
-- ============================================================================

-- Users table indexes
CREATE INDEX idx_users_email ON authentication.users(email);
CREATE INDEX idx_users_username ON authentication.users(username);
CREATE INDEX idx_users_phone ON authentication.users(phone);
CREATE INDEX idx_users_status ON authentication.users(status);
CREATE INDEX idx_users_created_at ON authentication.users(created_at);
CREATE INDEX idx_users_deleted_at ON authentication.users(deleted_at);

-- Profiles table indexes
CREATE UNIQUE INDEX idx_profiles_user_id ON authentication.profiles(user_id);
CREATE INDEX idx_profiles_display_name ON authentication.profiles(display_name);
CREATE INDEX idx_profiles_country ON authentication.profiles(country);
CREATE INDEX idx_profiles_created_at ON authentication.profiles(created_at);

-- Token access table indexes
CREATE INDEX idx_token_access_user_id ON authentication.token_access(user_id);
CREATE INDEX idx_token_access_token_hash ON authentication.token_access(token_hash);
CREATE INDEX idx_token_access_token_type ON authentication.token_access(token_type);
CREATE INDEX idx_token_access_expires_at ON authentication.token_access(expires_at);
CREATE INDEX idx_token_access_revoked ON authentication.token_access(revoked);
CREATE INDEX idx_token_access_created_at ON authentication.token_access(created_at);

-- Password resets table indexes
CREATE INDEX idx_password_resets_user_id ON authentication.password_resets(user_id);
CREATE INDEX idx_password_resets_email ON authentication.password_resets(email);
CREATE INDEX idx_password_resets_token_hash ON authentication.password_resets(token_hash);
CREATE INDEX idx_password_resets_expires_at ON authentication.password_resets(expires_at);
CREATE INDEX idx_password_resets_used ON authentication.password_resets(used);
CREATE INDEX idx_password_resets_created_at ON authentication.password_resets(created_at);

-- Permissions table indexes
CREATE INDEX idx_permissions_user_id ON authentication.permissions(user_id);
CREATE INDEX idx_permissions_role ON authentication.permissions(role);
CREATE INDEX idx_permissions_permission ON authentication.permissions(permission);
CREATE INDEX idx_permissions_resource ON authentication.permissions(resource);
CREATE INDEX idx_permissions_resource_id ON authentication.permissions(resource_id);
CREATE INDEX idx_permissions_expires_at ON authentication.permissions(expires_at);
CREATE INDEX idx_permissions_revoked ON authentication.permissions(revoked);
CREATE INDEX idx_permissions_created_at ON authentication.permissions(created_at);
CREATE INDEX idx_permissions_user_role ON authentication.permissions(user_id, role);
CREATE INDEX idx_permissions_user_permission ON authentication.permissions(user_id, permission);
CREATE INDEX idx_permissions_role_permission ON authentication.permissions(role, permission);

-- ============================================================================
-- Create Functions and Triggers
-- ============================================================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION authentication.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers
CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON authentication.users 
    FOR EACH ROW EXECUTE FUNCTION authentication.update_updated_at_column();

CREATE TRIGGER update_profiles_updated_at 
    BEFORE UPDATE ON authentication.profiles 
    FOR EACH ROW EXECUTE FUNCTION authentication.update_updated_at_column();

CREATE TRIGGER update_token_access_updated_at 
    BEFORE UPDATE ON authentication.token_access 
    FOR EACH ROW EXECUTE FUNCTION authentication.update_updated_at_column();

CREATE TRIGGER update_password_resets_updated_at 
    BEFORE UPDATE ON authentication.password_resets 
    FOR EACH ROW EXECUTE FUNCTION authentication.update_updated_at_column();

CREATE TRIGGER update_permissions_updated_at 
    BEFORE UPDATE ON authentication.permissions 
    FOR EACH ROW EXECUTE FUNCTION authentication.update_updated_at_column();

-- ============================================================================
-- Create Views
-- ============================================================================

-- Active users with profiles view
CREATE VIEW authentication.v_active_users_with_profiles AS
SELECT 
    u.id,
    u.username,
    u.email,
    u.email_verified_at,
    u.phone,
    u.phone_verified_at,
    u.status,
    u.last_login_at,
    u.two_factor_enabled,
    u.created_at as user_created_at,
    p.first_name,
    p.last_name,
    p.display_name,
    p.avatar_url,
    p.bio,
    p.date_of_birth,
    p.gender,
    p.country,
    p.city,
    p.timezone,
    p.language
FROM authentication.users u
LEFT JOIN authentication.profiles p ON u.id = p.user_id
WHERE u.status = 'active' AND u.deleted_at IS NULL;

-- User permissions view
CREATE VIEW authentication.v_user_permissions AS
SELECT 
    u.id as user_id,
    u.username,
    u.email,
    p.role,
    p.permission,
    p.resource,
    p.resource_id,
    p.expires_at,
    p.revoked,
    p.conditions,
    p.granted_at
FROM authentication.users u
JOIN authentication.permissions p ON u.id = p.user_id
WHERE u.status = 'active' 
  AND u.deleted_at IS NULL 
  AND p.revoked = FALSE 
  AND (p.expires_at IS NULL OR p.expires_at > CURRENT_TIMESTAMP);

-- ============================================================================
-- Add Comments
-- ============================================================================

COMMENT ON SCHEMA authentication IS 'Authentication and authorization schema for the application';
COMMENT ON TABLE authentication.users IS 'Core user authentication data including credentials and security settings';
COMMENT ON TABLE authentication.profiles IS 'Extended user profile information and preferences';
COMMENT ON TABLE authentication.token_access IS 'Access tokens, refresh tokens, and API keys management';
COMMENT ON TABLE authentication.password_resets IS 'Password reset tokens and history tracking';
COMMENT ON TABLE authentication.permissions IS 'Role-based access control permissions for users';

-- ============================================================================
-- Insert Initial Data (Optional)
-- ============================================================================

-- Create system admin user (password should be changed immediately)
-- Note: This is just an example - password should be properly hashed
-- INSERT INTO authentication.users (username, email, password_hash, status, email_verified_at) 
-- VALUES ('system_admin', 'admin@yourdomain.com', '$2a$10$example_hash_replace_with_real_hash', 'active', CURRENT_TIMESTAMP);

-- Create system admin profile
-- INSERT INTO authentication.profiles (user_id, first_name, last_name, display_name) 
-- VALUES ((SELECT id FROM authentication.users WHERE username = 'system_admin'), 'System', 'Administrator', 'System Admin');

-- Grant super admin permissions
-- INSERT INTO authentication.permissions (user_id, role, permission) 
-- VALUES ((SELECT id FROM authentication.users WHERE username = 'system_admin'), 'super_admin', '*');

COMMIT;

-- ============================================================================
-- Down Migration (for rollback)
-- ============================================================================

-- To rollback this migration, run:
-- BEGIN;
-- DROP SCHEMA IF EXISTS authentication CASCADE;
-- COMMIT;
