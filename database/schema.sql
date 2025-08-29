-- ============================================================================
-- Authentication Database Schema for PostgreSQL
-- Created: August 23, 2025
-- Description: Complete authentication system schema with users, profiles, 
--              tokens, password resets, and permissions
-- ============================================================================

-- Create authentication schema
CREATE SCHEMA IF NOT EXISTS authentication;

-- Set default schema for current session
SET search_path TO authentication, public;

-- ============================================================================
-- 1. USERS TABLE
-- Core user authentication data
-- ============================================================================
CREATE TABLE IF NOT EXISTS authentication.users (
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

-- Indexes for users table
CREATE INDEX IF NOT EXISTS idx_users_email ON authentication.users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON authentication.users(username);
CREATE INDEX IF NOT EXISTS idx_users_phone ON authentication.users(phone);
CREATE INDEX IF NOT EXISTS idx_users_status ON authentication.users(status);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON authentication.users(created_at);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON authentication.users(deleted_at);

-- ============================================================================
-- 2. PROFILES TABLE
-- Extended user profile information
-- ============================================================================
CREATE TABLE IF NOT EXISTS authentication.profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES authentication.users(id) ON DELETE CASCADE,
    first_name VARCHAR(100) NULL,
    last_name VARCHAR(100) NULL,
    display_name VARCHAR(200) NULL,
    avatar_url VARCHAR(500) NULL,
    bio TEXT NULL,
    date_of_birth DATE NULL,
    gender VARCHAR(10) CHECK (gender IN ('male', 'female', 'other', 'prefer_not_to_say')),
    country VARCHAR(2) NULL, -- ISO 3166-1 alpha-2 country code
    state VARCHAR(100) NULL,
    city VARCHAR(100) NULL,
    address TEXT NULL,
    postal_code VARCHAR(20) NULL,
    timezone VARCHAR(50) DEFAULT 'UTC',
    language VARCHAR(5) DEFAULT 'en', -- ISO 639-1 language code
    website_url VARCHAR(500) NULL,
    social_links JSONB NULL, -- Store social media links as JSON
    preferences JSONB NULL, -- Store user preferences as JSON
    metadata JSONB NULL, -- Additional flexible data storage
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for profiles table
CREATE UNIQUE INDEX IF NOT EXISTS idx_profiles_user_id ON authentication.profiles(user_id);
CREATE INDEX IF NOT EXISTS idx_profiles_display_name ON authentication.profiles(display_name);
CREATE INDEX IF NOT EXISTS idx_profiles_country ON authentication.profiles(country);
CREATE INDEX IF NOT EXISTS idx_profiles_created_at ON authentication.profiles(created_at);

-- ============================================================================
-- 3. TOKEN_ACCESS TABLE
-- Access tokens, refresh tokens, and API keys management
-- ============================================================================
CREATE TABLE IF NOT EXISTS authentication.token_access (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES authentication.users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL, -- Hashed token value
    token_type VARCHAR(20) NOT NULL CHECK (token_type IN ('access', 'refresh', 'api_key', 'verification')),
    token_name VARCHAR(100) NULL, -- Optional name for API keys
    scope TEXT NULL, -- Comma-separated list of scopes/permissions
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    last_used_at TIMESTAMP WITH TIME ZONE NULL,
    last_used_ip INET NULL,
    user_agent TEXT NULL,
    device_info JSONB NULL, -- Store device information as JSON
    revoked BOOLEAN DEFAULT FALSE,
    revoked_at TIMESTAMP WITH TIME ZONE NULL,
    revoked_by UUID NULL REFERENCES authentication.users(id),
    revoke_reason VARCHAR(255) NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for token_access table
CREATE INDEX IF NOT EXISTS idx_token_access_user_id ON authentication.token_access(user_id);
CREATE INDEX IF NOT EXISTS idx_token_access_token_hash ON authentication.token_access(token_hash);
CREATE INDEX IF NOT EXISTS idx_token_access_token_type ON authentication.token_access(token_type);
CREATE INDEX IF NOT EXISTS idx_token_access_expires_at ON authentication.token_access(expires_at);
CREATE INDEX IF NOT EXISTS idx_token_access_revoked ON authentication.token_access(revoked);
CREATE INDEX IF NOT EXISTS idx_token_access_created_at ON authentication.token_access(created_at);

-- ============================================================================
-- 4. PASSWORD_RESETS TABLE
-- Password reset tokens and history
-- ============================================================================
CREATE TABLE IF NOT EXISTS authentication.password_resets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES authentication.users(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    token_hash VARCHAR(255) NOT NULL, -- Hashed reset token
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    used_at TIMESTAMP WITH TIME ZONE NULL,
    used_ip INET NULL,
    user_agent TEXT NULL,
    attempts INTEGER DEFAULT 0, -- Number of verification attempts
    max_attempts INTEGER DEFAULT 3,
    blocked_until TIMESTAMP WITH TIME ZONE NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for password_resets table
CREATE INDEX IF NOT EXISTS idx_password_resets_user_id ON authentication.password_resets(user_id);
CREATE INDEX IF NOT EXISTS idx_password_resets_email ON authentication.password_resets(email);
CREATE INDEX IF NOT EXISTS idx_password_resets_token_hash ON authentication.password_resets(token_hash);
CREATE INDEX IF NOT EXISTS idx_password_resets_expires_at ON authentication.password_resets(expires_at);
CREATE INDEX IF NOT EXISTS idx_password_resets_used ON authentication.password_resets(used);
CREATE INDEX IF NOT EXISTS idx_password_resets_created_at ON authentication.password_resets(created_at);

-- ============================================================================
-- 5. PERMISSIONS TABLE
-- Role-based access control (RBAC) permissions
-- ============================================================================
CREATE TABLE IF NOT EXISTS authentication.permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES authentication.users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL,
    permission VARCHAR(100) NOT NULL,
    resource VARCHAR(100) NULL, -- Optional specific resource
    resource_id UUID NULL, -- Optional specific resource ID
    granted_by UUID NULL REFERENCES authentication.users(id),
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NULL, -- NULL means no expiration
    revoked BOOLEAN DEFAULT FALSE,
    revoked_at TIMESTAMP WITH TIME ZONE NULL,
    revoked_by UUID NULL REFERENCES authentication.users(id),
    conditions JSONB NULL, -- Additional conditions for permission (time-based, IP-based, etc.)
    metadata JSONB NULL, -- Additional metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for permissions table
CREATE INDEX IF NOT EXISTS idx_permissions_user_id ON authentication.permissions(user_id);
CREATE INDEX IF NOT EXISTS idx_permissions_role ON authentication.permissions(role);
CREATE INDEX IF NOT EXISTS idx_permissions_permission ON authentication.permissions(permission);
CREATE INDEX IF NOT EXISTS idx_permissions_resource ON authentication.permissions(resource);
CREATE INDEX IF NOT EXISTS idx_permissions_resource_id ON authentication.permissions(resource_id);
CREATE INDEX IF NOT EXISTS idx_permissions_expires_at ON authentication.permissions(expires_at);
CREATE INDEX IF NOT EXISTS idx_permissions_revoked ON authentication.permissions(revoked);
CREATE INDEX IF NOT EXISTS idx_permissions_created_at ON authentication.permissions(created_at);

-- Composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_permissions_user_role ON authentication.permissions(user_id, role);
CREATE INDEX IF NOT EXISTS idx_permissions_user_permission ON authentication.permissions(user_id, permission);
CREATE INDEX IF NOT EXISTS idx_permissions_role_permission ON authentication.permissions(role, permission);

-- ============================================================================
-- TRIGGERS FOR UPDATED_AT TIMESTAMPS
-- ============================================================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION authentication.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply trigger to all tables with updated_at column
CREATE OR REPLACE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON authentication.users 
    FOR EACH ROW EXECUTE FUNCTION authentication.update_updated_at_column();

CREATE OR REPLACE TRIGGER update_profiles_updated_at 
    BEFORE UPDATE ON authentication.profiles 
    FOR EACH ROW EXECUTE FUNCTION authentication.update_updated_at_column();

CREATE OR REPLACE TRIGGER update_token_access_updated_at 
    BEFORE UPDATE ON authentication.token_access 
    FOR EACH ROW EXECUTE FUNCTION authentication.update_updated_at_column();

CREATE OR REPLACE TRIGGER update_password_resets_updated_at 
    BEFORE UPDATE ON authentication.password_resets 
    FOR EACH ROW EXECUTE FUNCTION authentication.update_updated_at_column();

CREATE OR REPLACE TRIGGER update_permissions_updated_at 
    BEFORE UPDATE ON authentication.permissions 
    FOR EACH ROW EXECUTE FUNCTION authentication.update_updated_at_column();

-- ============================================================================
-- VIEWS FOR COMMON QUERIES
-- ============================================================================

-- View for active users with profiles
CREATE OR REPLACE VIEW authentication.v_active_users_with_profiles AS
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

-- View for user permissions
CREATE OR REPLACE VIEW authentication.v_user_permissions AS
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
-- SAMPLE DATA INSERTION (Optional - for testing)
-- ============================================================================

-- Insert sample admin user (password: 'admin123' - should be properly hashed in production)
-- INSERT INTO authentication.users (username, email, password_hash, status, email_verified_at) 
-- VALUES ('admin', 'admin@example.com', '$2a$10$example_hash_here', 'active', CURRENT_TIMESTAMP);

-- Insert sample profile
-- INSERT INTO authentication.profiles (user_id, first_name, last_name, display_name) 
-- VALUES ((SELECT id FROM authentication.users WHERE username = 'admin'), 'System', 'Administrator', 'Admin');

-- Insert sample permissions
-- INSERT INTO authentication.permissions (user_id, role, permission) 
-- VALUES ((SELECT id FROM authentication.users WHERE username = 'admin'), 'super_admin', '*');

-- ============================================================================
-- COMMENTS FOR DOCUMENTATION
-- ============================================================================

COMMENT ON SCHEMA authentication IS 'Authentication and authorization schema for the application';

COMMENT ON TABLE authentication.users IS 'Core user authentication data including credentials and security settings';
COMMENT ON TABLE authentication.profiles IS 'Extended user profile information and preferences';
COMMENT ON TABLE authentication.token_access IS 'Access tokens, refresh tokens, and API keys management';
COMMENT ON TABLE authentication.password_resets IS 'Password reset tokens and history tracking';
COMMENT ON TABLE authentication.permissions IS 'Role-based access control permissions for users';

COMMENT ON COLUMN authentication.users.password_hash IS 'Bcrypt or Argon2 hashed password';
COMMENT ON COLUMN authentication.users.two_factor_secret IS 'TOTP secret for two-factor authentication';
COMMENT ON COLUMN authentication.profiles.social_links IS 'JSON object containing social media profile links';
COMMENT ON COLUMN authentication.profiles.preferences IS 'JSON object containing user preferences and settings';
COMMENT ON COLUMN authentication.token_access.scope IS 'Comma-separated list of permissions/scopes for this token';
COMMENT ON COLUMN authentication.permissions.conditions IS 'JSON object with additional permission conditions (time, IP, etc.)';
