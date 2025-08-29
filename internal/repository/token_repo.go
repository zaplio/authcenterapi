package repository

import (
	"authcenterapi/internal/provider"
	model "authcenterapi/model/entity"
	"context"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type tokenRepo struct {
	logger provider.ILogger
	conn   *pgx.Conn
}

// NewTokenRepository creates a new token repository
func NewTokenRepository(logger provider.ILogger, conn *pgx.Conn) TokenRepository {
	return &tokenRepo{
		logger: logger,
		conn:   conn,
	}
}

const (
	insertTokenQuery = `
		INSERT INTO authentication.token_access 
		(id, user_id, token_hash, token_type, device_info, last_used_ip, user_agent, 
		 expires_at, last_used_at, revoked_at, revoked_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	getTokenByHashQuery = `
		SELECT id, user_id, token_hash, token_type, device_info, last_used_ip, user_agent,
			   expires_at, last_used_at, revoked_at, revoked_by, created_at, updated_at
		FROM authentication.token_access 
		WHERE token_hash = $1 AND revoked_at IS NULL`

	getTokenWithUserQuery = `
		SELECT t.id, t.user_id, t.token_hash, t.token_type, t.device_info, t.last_used_ip, 
			   t.user_agent, t.expires_at, t.last_used_at, t.revoked_at, t.revoked_by, 
			   t.created_at, t.updated_at,
			   u.id, u.username, u.email, u.email_verified_at, u.phone, u.phone_verified_at, 
			   u.status, u.last_login_at, u.last_login_ip, u.failed_login_attempts, u.locked_until, 
			   u.two_factor_enabled, u.two_factor_secret, u.created_at, u.updated_at, u.deleted_at
		FROM authentication.token_access t
		INNER JOIN authentication.users u ON t.user_id = u.id
		WHERE t.token_hash = $1 AND t.revoked_at IS NULL AND u.deleted_at IS NULL`

	updateLastUsedQuery = `
		UPDATE authentication.token_access 
		SET last_used_at = $2, updated_at = $3
		WHERE id = $1`

	revokeTokenQuery = `
		UPDATE authentication.token_access 
		SET revoked_at = $2, revoked_by = $3, updated_at = $2
		WHERE id = $1`

	revokeAllUserTokensQuery = `
		UPDATE authentication.token_access 
		SET revoked_at = $2, revoked_by = $3, updated_at = $2
		WHERE user_id = $1 AND revoked_at IS NULL`

	getTokensByUserQuery = `
		SELECT id, user_id, token_hash, token_type, device_info, last_used_ip, user_agent,
			   expires_at, last_used_at, revoked_at, revoked_by, created_at, updated_at
		FROM authentication.token_access 
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`
)

func (r *tokenRepo) Create(ctx context.Context, token *model.TokenAccess) error {
	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}

	now := time.Now()
	token.CreatedAt = now
	token.UpdatedAt = now

	_, err := r.conn.Exec(ctx, insertTokenQuery,
		token.ID, token.UserID, token.TokenHash, token.TokenType, token.DeviceInfo,
		token.LastUsedIP, token.UserAgent, token.ExpiresAt, token.LastUsedAt,
		token.RevokedAt, token.RevokedBy, token.CreatedAt, token.UpdatedAt)

	if err != nil {
		r.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to create token error %v, user_id %s, token_type %s", err, token.UserID, token.TokenType)
		return err
	}

	r.logger.Infofctx(provider.AppLog, ctx, "Token created successfully token_id %s, user_id %s, token_type %s", token.ID, token.UserID, token.TokenType)
	return nil
}

func (r *tokenRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*model.TokenAccess, error) {
	token := &model.TokenAccess{}

	err := r.conn.QueryRow(ctx, getTokenByHashQuery, tokenHash).Scan(
		&token.ID, &token.UserID, &token.TokenHash, &token.TokenType, &token.DeviceInfo,
		&token.LastUsedIP, &token.UserAgent, &token.ExpiresAt, &token.LastUsedAt,
		&token.RevokedAt, &token.RevokedBy, &token.CreatedAt, &token.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to get token by hash, caused by %v", err)
		return nil, err
	}

	return token, nil
}

func (r *tokenRepo) GetTokenWithUser(ctx context.Context, tokenHash string) (*model.TokenAccess, *model.User, error) {
	token := &model.TokenAccess{}
	user := &model.User{}

	err := r.conn.QueryRow(ctx, getTokenWithUserQuery, tokenHash).Scan(
		&token.ID, &token.UserID, &token.TokenHash, &token.TokenType, &token.DeviceInfo,
		&token.LastUsedIP, &token.UserAgent, &token.ExpiresAt, &token.LastUsedAt,
		&token.RevokedAt, &token.RevokedBy, &token.CreatedAt, &token.UpdatedAt,
		&user.ID, &user.Username, &user.Email, &user.EmailVerifiedAt, &user.Phone,
		&user.PhoneVerifiedAt, &user.Status, &user.LastLoginAt, &user.LastLoginIP,
		&user.FailedLoginAttempts, &user.LockedUntil, &user.TwoFactorEnabled,
		&user.TwoFactorSecret, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil, nil
		}
		r.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to get token with user, caused by %v", err)
		return nil, nil, err
	}

	return token, user, nil
}

func (r *tokenRepo) UpdateLastUsed(ctx context.Context, tokenID uuid.UUID, ipAddress net.IP) error {
	now := time.Now()

	_, err := r.conn.Exec(ctx, updateLastUsedQuery, tokenID, now, now)
	if err != nil {
		r.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to update token last used token_id %s ip %s, caused by %v", tokenID, ipAddress, err)
		return err
	}

	r.logger.Debugfctx(provider.AppLog, ctx, "Token last used updated token_id %s, ip %s", tokenID, ipAddress)
	return nil
}

func (r *tokenRepo) RevokeToken(ctx context.Context, tokenID uuid.UUID, revokedBy *uuid.UUID, reason string) error {
	now := time.Now()

	_, err := r.conn.Exec(ctx, revokeTokenQuery, tokenID, now, revokedBy)
	if err != nil {
		r.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to revoke token token_id %s, caused by %v", tokenID, err)
		return err
	}

	r.logger.Infofctx(provider.AppLog, ctx, "Token revoked successfully token_id %s, revoked_by %s, reason %s", tokenID, revokedBy, reason)
	return nil
}

func (r *tokenRepo) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID, revokedBy *uuid.UUID, reason string) error {
	now := time.Now()

	result, err := r.conn.Exec(ctx, revokeAllUserTokensQuery, userID, now, revokedBy)
	if err != nil {
		r.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to revoke all user tokens for user_id %s, caused by %v", userID, err)
		return err
	}

	rowsAffected := result.RowsAffected()
	r.logger.Infofctx(provider.AppLog, ctx, "All user tokens revoked for user_id %s, revoked_by %s, reason %s, tokens_revoked %d", userID, revokedBy, reason, rowsAffected)
	return nil
}

func (r *tokenRepo) GetTokensByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.TokenAccess, error) {
	rows, err := r.conn.Query(ctx, getTokensByUserQuery, userID, limit, offset)
	if err != nil {
		r.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to get tokens by user_id %s, caused by %v", userID, err)
		return nil, err
	}
	defer rows.Close()

	var tokens []*model.TokenAccess
	for rows.Next() {
		token := &model.TokenAccess{}
		err := rows.Scan(
			&token.ID, &token.UserID, &token.TokenHash, &token.TokenType, &token.DeviceInfo,
			&token.LastUsedIP, &token.UserAgent, &token.ExpiresAt, &token.LastUsedAt,
			&token.RevokedAt, &token.RevokedBy, &token.CreatedAt, &token.UpdatedAt)

		if err != nil {
			r.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to scan token row, caused by %v", err)
			return nil, err
		}

		tokens = append(tokens, token)
	}

	if err := rows.Err(); err != nil {
		r.logger.Errorfctx(provider.AppLog, ctx, false, "Row iteration error, caused by %v", err)
		return nil, err
	}

	return tokens, nil
}
