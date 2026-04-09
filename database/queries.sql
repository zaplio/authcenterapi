-- ============================================================================
-- Authentication System - Common SQL Queries
-- Created: August 23, 2025
-- Description: Common queries for authentication system operations
-- ============================================================================

-- Set schema path
SET search_path TO authentication, public;

-- ============================================================================
-- USER MANAGEMENT QUERIES
-- ============================================================================

-- 1. Create a new user with profile
INSERT INTO authentication.users (username, email, password_hash, phone) 
VALUES ('john_doe', 'john@example.com', '$2a$10$hashed_password_here', '+1234567890')
RETURNING id, username, email, created_at;

-- 2. Get user by email or username
SELECT id, username, email, status, last_login_at, two_factor_enabled, created_at
FROM authentication.users 
WHERE (email = $1 OR username = $1) AND deleted_at IS NULL;

-- 3. Update user profile
UPDATE authentication.profiles 
SET first_name = $1, last_name = $2, display_name = $3, bio = $4, updated_at = CURRENT_TIMESTAMP
WHERE user_id = $5;

-- 4. Get user with complete profile
SELECT 
    u.id, u.username, u.email, u.phone, u.status, u.last_login_at, u.two_factor_enabled,
    p.first_name, p.last_name, p.display_name, p.avatar_url, p.bio, p.country, p.city, p.timezone
FROM authentication.users u
LEFT JOIN authentication.profiles p ON u.id = p.user_id
WHERE u.id = $1 AND u.deleted_at IS NULL;

-- 5. Search users by name or email
SELECT u.id, u.username, u.email, p.first_name, p.last_name, p.display_name
FROM authentication.users u
LEFT JOIN authentication.profiles p ON u.id = p.user_id
WHERE u.deleted_at IS NULL 
  AND (
    LOWER(u.username) LIKE LOWER($1) OR 
    LOWER(u.email) LIKE LOWER($1) OR 
    LOWER(p.display_name) LIKE LOWER($1) OR
    LOWER(CONCAT(p.first_name, ' ', p.last_name)) LIKE LOWER($1)
  )
ORDER BY u.created_at DESC
LIMIT 20;

-- ============================================================================
-- AUTHENTICATION QUERIES
-- ============================================================================

-- 6. Validate user login credentials
SELECT id, username, email, password_hash, status, failed_login_attempts, locked_until
FROM authentication.users 
WHERE (email = $1 OR username = $1) 
  AND status = 'active' 
  AND deleted_at IS NULL;

-- 7. Update login information after successful authentication
UPDATE authentication.users 
SET last_login_at = CURRENT_TIMESTAMP, 
    last_login_ip = $2, 
    failed_login_attempts = 0,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- 8. Increment failed login attempts
UPDATE authentication.users 
SET failed_login_attempts = failed_login_attempts + 1,
    locked_until = CASE 
        WHEN failed_login_attempts + 1 >= 5 THEN CURRENT_TIMESTAMP + INTERVAL '30 minutes'
        ELSE locked_until
    END,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- 9. Check if user account is locked
SELECT id, locked_until 
FROM authentication.users 
WHERE id = $1 
  AND locked_until IS NOT NULL 
  AND locked_until > CURRENT_TIMESTAMP;

-- ============================================================================
-- TOKEN MANAGEMENT QUERIES
-- ============================================================================

-- 10. Create access token
INSERT INTO authentication.token_access 
(user_id, token_hash, token_type, scope, expires_at, user_agent, device_info)
VALUES ($1, $2, 'access', $3, $4, $5, $6)
RETURNING id, expires_at;

-- 11. Create refresh token
INSERT INTO authentication.token_access 
(user_id, token_hash, token_type, expires_at)
VALUES ($1, $2, 'refresh', $3)
RETURNING id, expires_at;

-- 12. Validate and get token information
SELECT t.id, t.user_id, t.token_type, t.scope, t.expires_at, t.revoked,
       u.username, u.email, u.status
FROM authentication.token_access t
JOIN authentication.users u ON t.user_id = u.id
WHERE t.token_hash = $1 
  AND t.revoked = FALSE 
  AND t.expires_at > CURRENT_TIMESTAMP
  AND u.status = 'active' 
  AND u.deleted_at IS NULL;

-- 13. Update token last used information
UPDATE authentication.token_access 
SET last_used_at = CURRENT_TIMESTAMP, 
    last_used_ip = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- 14. Revoke token
UPDATE authentication.token_access 
SET revoked = TRUE, 
    revoked_at = CURRENT_TIMESTAMP, 
    revoked_by = $2,
    revoke_reason = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- 15. Revoke all user tokens
UPDATE authentication.token_access 
SET revoked = TRUE, 
    revoked_at = CURRENT_TIMESTAMP, 
    revoked_by = $2,
    revoke_reason = 'User logout - all tokens revoked',
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = $1 AND revoked = FALSE;

-- 16. Clean up expired tokens
DELETE FROM authentication.token_access 
WHERE expires_at < CURRENT_TIMESTAMP - INTERVAL '30 days';

-- ============================================================================
-- PASSWORD RESET QUERIES
-- ============================================================================

-- 17. Create password reset request
INSERT INTO authentication.password_resets 
(user_id, email, token_hash, expires_at)
VALUES ($1, $2, $3, CURRENT_TIMESTAMP + INTERVAL '1 hour')
RETURNING id, expires_at;

-- 18. Validate password reset token
SELECT pr.id, pr.user_id, pr.email, pr.expires_at, pr.used, pr.attempts, pr.max_attempts, pr.blocked_until
FROM authentication.password_resets pr
WHERE pr.token_hash = $1 
  AND pr.used = FALSE 
  AND pr.expires_at > CURRENT_TIMESTAMP
  AND (pr.blocked_until IS NULL OR pr.blocked_until < CURRENT_TIMESTAMP);

-- 19. Increment reset attempts
UPDATE authentication.password_resets 
SET attempts = attempts + 1,
    blocked_until = CASE 
        WHEN attempts + 1 >= max_attempts THEN CURRENT_TIMESTAMP + INTERVAL '1 hour'
        ELSE blocked_until
    END,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- 20. Mark password reset as used
UPDATE authentication.password_resets 
SET used = TRUE, 
    used_at = CURRENT_TIMESTAMP, 
    used_ip = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- 21. Update user password
UPDATE authentication.users 
SET password_hash = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- ============================================================================
-- PERMISSION MANAGEMENT QUERIES
-- ============================================================================

-- 22. Grant permission to user
INSERT INTO authentication.permissions 
(user_id, role, permission, resource, granted_by)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, granted_at;

-- 23. Check user permission
SELECT COUNT(*) > 0 as has_permission
FROM authentication.permissions p
JOIN authentication.users u ON p.user_id = u.id
WHERE p.user_id = $1 
  AND (p.role = $2 OR p.permission = $3)
  AND p.revoked = FALSE
  AND (p.expires_at IS NULL OR p.expires_at > CURRENT_TIMESTAMP)
  AND u.status = 'active' 
  AND u.deleted_at IS NULL;

-- 24. Get all user permissions
SELECT role, permission, resource, resource_id, expires_at, granted_at
FROM authentication.permissions 
WHERE user_id = $1 
  AND revoked = FALSE 
  AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
ORDER BY granted_at DESC;

-- 25. Get users by role
SELECT u.id, u.username, u.email, p.first_name, p.last_name, p.display_name
FROM authentication.users u
LEFT JOIN authentication.profiles p ON u.id = p.user_id
JOIN authentication.permissions perm ON u.id = perm.user_id
WHERE perm.role = $1 
  AND perm.revoked = FALSE 
  AND (perm.expires_at IS NULL OR perm.expires_at > CURRENT_TIMESTAMP)
  AND u.status = 'active' 
  AND u.deleted_at IS NULL
GROUP BY u.id, u.username, u.email, p.first_name, p.last_name, p.display_name
ORDER BY u.created_at DESC;

-- 26. Revoke permission
UPDATE authentication.permissions 
SET revoked = TRUE, 
    revoked_at = CURRENT_TIMESTAMP, 
    revoked_by = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- ============================================================================
-- REPORTING AND ANALYTICS QUERIES
-- ============================================================================

-- 27. Get user statistics
SELECT 
    COUNT(*) as total_users,
    COUNT(CASE WHEN status = 'active' THEN 1 END) as active_users,
    COUNT(CASE WHEN email_verified_at IS NOT NULL THEN 1 END) as verified_users,
    COUNT(CASE WHEN two_factor_enabled = TRUE THEN 1 END) as two_factor_users,
    COUNT(CASE WHEN last_login_at > CURRENT_TIMESTAMP - INTERVAL '30 days' THEN 1 END) as active_last_30_days
FROM authentication.users 
WHERE deleted_at IS NULL;

-- 28. Get login activity (last 7 days)
SELECT 
    DATE(last_login_at) as login_date,
    COUNT(DISTINCT id) as unique_logins
FROM authentication.users 
WHERE last_login_at > CURRENT_TIMESTAMP - INTERVAL '7 days'
  AND deleted_at IS NULL
GROUP BY DATE(last_login_at)
ORDER BY login_date DESC;

-- 29. Get token usage statistics
SELECT 
    token_type,
    COUNT(*) as total_tokens,
    COUNT(CASE WHEN revoked = FALSE THEN 1 END) as active_tokens,
    COUNT(CASE WHEN last_used_at > CURRENT_TIMESTAMP - INTERVAL '24 hours' THEN 1 END) as used_last_24h
FROM authentication.token_access 
GROUP BY token_type;

-- 30. Get most common user countries
SELECT 
    p.country,
    COUNT(*) as user_count
FROM authentication.profiles p
JOIN authentication.users u ON p.user_id = u.id
WHERE u.status = 'active' AND u.deleted_at IS NULL AND p.country IS NOT NULL
GROUP BY p.country
ORDER BY user_count DESC
LIMIT 10;

-- ============================================================================
-- MAINTENANCE QUERIES
-- ============================================================================

-- 31. Clean up old password reset tokens
DELETE FROM authentication.password_resets 
WHERE created_at < CURRENT_TIMESTAMP - INTERVAL '7 days';

-- 32. Find users with multiple failed login attempts
SELECT id, username, email, failed_login_attempts, locked_until, last_login_at
FROM authentication.users 
WHERE failed_login_attempts > 3 
  AND status = 'active' 
  AND deleted_at IS NULL
ORDER BY failed_login_attempts DESC;

-- 33. Find inactive users (no login in 90 days)
SELECT u.id, u.username, u.email, u.last_login_at, u.created_at
FROM authentication.users u
WHERE (u.last_login_at IS NULL OR u.last_login_at < CURRENT_TIMESTAMP - INTERVAL '90 days')
  AND u.status = 'active' 
  AND u.deleted_at IS NULL
  AND u.created_at < CURRENT_TIMESTAMP - INTERVAL '90 days'
ORDER BY u.last_login_at ASC NULLS FIRST;

-- 34. Soft delete user (instead of hard delete)
UPDATE authentication.users 
SET status = 'deleted', 
    deleted_at = CURRENT_TIMESTAMP, 
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- 35. Restore soft deleted user
UPDATE authentication.users 
SET status = 'active', 
    deleted_at = NULL, 
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;
