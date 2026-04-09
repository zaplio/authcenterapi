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

type passwordResetRepo struct {
	logger provider.ILogger
	conn   *pgx.Conn
}

// NewPasswordResetRepository creates a new password reset repository
func NewPasswordResetRepository(logger provider.ILogger, conn *pgx.Conn) PasswordResetRepository {
	return &passwordResetRepo{
		logger: logger,
		conn:   conn,
	}
}

const (
	insertPasswordResetQuery = `
		INSERT INTO authentication.password_resets 
		(id, user_id, email, token_hash, expires_at, used_at, used_ip, user_agent, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	getPasswordResetByTokenQuery = `
		SELECT id, user_id, token_hash, expires_at, used_at, used_ip, user_agent, created_at, updated_at
		FROM authentication.password_resets 
		WHERE token_hash = $1 AND used_at IS NULL AND expires_at > $2`

	getPasswordResetWithUserQuery = `
		SELECT pr.id, pr.user_id, pr.token_hash, pr.expires_at, pr.used_at, pr.used_ip, 
			   pr.user_agent, pr.created_at, pr.updated_at,
			   u.id, u.username, u.email, u.email_verified_at, u.phone, u.phone_verified_at, 
			   u.status, u.last_login_at, u.last_login_ip, u.failed_login_attempts, u.locked_until, 
			   u.two_factor_enabled, u.two_factor_secret, u.created_at, u.updated_at, u.deleted_at
		FROM authentication.password_resets pr
		INNER JOIN authentication.users u ON pr.user_id = u.id
		WHERE pr.token_hash = $1 AND pr.used_at IS NULL AND pr.expires_at > $2 AND u.deleted_at IS NULL`

	markPasswordResetAsUsedQuery = `
		UPDATE authentication.password_resets 
		SET used_at = $2, updated_at = $2
		WHERE id = $1`

	deleteExpiredPasswordResetsQuery = `
		DELETE FROM authentication.password_resets 
		WHERE expires_at < $1`

	getUserPasswordResetsQuery = `
		SELECT id, user_id, token_hash, expires_at, used_at, ip_address, user_agent, created_at, updated_at
		FROM authentication.password_resets 
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	revokeUserPasswordResetsQuery = `
		UPDATE authentication.password_resets 
		SET used_at = $2, updated_at = $2
		WHERE user_id = $1 AND used_at IS NULL`
)

func (r *passwordResetRepo) Create(ctx context.Context, passwordReset *model.PasswordReset) error {
	if passwordReset.ID == uuid.Nil {
		passwordReset.ID = uuid.New()
	}

	now := time.Now()
	passwordReset.CreatedAt = now
	passwordReset.UpdatedAt = now

	_, err := r.conn.Exec(ctx, insertPasswordResetQuery,
		passwordReset.ID, passwordReset.UserID, passwordReset.Email, passwordReset.TokenHash,
		passwordReset.ExpiresAt, passwordReset.UsedAt, passwordReset.UsedIP,
		passwordReset.UserAgent, passwordReset.CreatedAt, passwordReset.UpdatedAt)

	if err != nil {
		r.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to create password reset error %v, user_id %s", err, passwordReset.UserID)
		return err
	}

	r.logger.Infofctx(provider.AppLog, ctx, "Password reset created successfully reset_id %s, user_id %s", passwordReset.ID, passwordReset.UserID)
	return nil
}

func (r *passwordResetRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*model.PasswordReset, error) {
	now := time.Now()
	passwordReset := &model.PasswordReset{}

	err := r.conn.QueryRow(ctx, getPasswordResetByTokenQuery, tokenHash, now).Scan(
		&passwordReset.ID, &passwordReset.UserID, &passwordReset.TokenHash,
		&passwordReset.ExpiresAt, &passwordReset.UsedAt, &passwordReset.UsedIP,
		&passwordReset.UserAgent, &passwordReset.CreatedAt, &passwordReset.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to get password reset by token, caused by %v", err)
		return nil, err
	}

	return passwordReset, nil
}

func (r *passwordResetRepo) GetPasswordResetWithUser(ctx context.Context, tokenHash string) (*model.PasswordResetWithUser, error) {
	now := time.Now()
	passwordResetWithUser := &model.PasswordResetWithUser{}
	passwordReset := &passwordResetWithUser.PasswordReset
	user := &passwordResetWithUser.User

	err := r.conn.QueryRow(ctx, getPasswordResetWithUserQuery, tokenHash, now).Scan(
		&passwordReset.ID, &passwordReset.UserID, &passwordReset.TokenHash,
		&passwordReset.ExpiresAt, &passwordReset.UsedAt, &passwordReset.UsedIP,
		&passwordReset.UserAgent, &passwordReset.CreatedAt, &passwordReset.UpdatedAt,
		&user.ID, &user.Username, &user.Email, &user.EmailVerifiedAt, &user.Phone,
		&user.PhoneVerifiedAt, &user.Status, &user.LastLoginAt, &user.LastLoginIP,
		&user.FailedLoginAttempts, &user.LockedUntil, &user.TwoFactorEnabled,
		&user.TwoFactorSecret, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to get password reset with user, caused by %v", err)
		return nil, err
	}

	return passwordResetWithUser, nil
}

func (r *passwordResetRepo) MarkAsUsed(ctx context.Context, passwordResetID uuid.UUID, ipAddress net.IP, userAgent string) error {
	now := time.Now()

	query := `UPDATE authentication.password_resets SET used_at = $2, used_ip = $3, user_agent = $4, updated_at = $2 WHERE id = $1`
	_, err := r.conn.Exec(ctx, query, passwordResetID, now, ipAddress, userAgent)
	if err != nil {
		r.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to mark password reset as used, caused by %v", err)
		return err
	}

	r.logger.Infofctx(provider.AppLog, ctx, "Password reset marked as used, reset_id %s", passwordResetID)
	return nil
}

func (r *passwordResetRepo) DeleteExpired(ctx context.Context) (int64, error) {
	now := time.Now()

	result, err := r.conn.Exec(ctx, deleteExpiredPasswordResetsQuery, now)
	if err != nil {
		r.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to delete expired password resets, caused by %v", err)
		return 0, err
	}

	rowsAffected := result.RowsAffected()
	r.logger.Infofctx(provider.AppLog, ctx, "Expired password resets deleted, resets_deleted %d", rowsAffected)
	return rowsAffected, nil
}

func (r *passwordResetRepo) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.PasswordReset, error) {
	rows, err := r.conn.Query(ctx, getUserPasswordResetsQuery, userID, limit, offset)
	if err != nil {
		r.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to get password resets by user, caused by %v", err)
		return nil, err
	}
	defer rows.Close()

	var passwordResets []*model.PasswordReset
	for rows.Next() {
		passwordReset := &model.PasswordReset{}
		err := rows.Scan(
			&passwordReset.ID, &passwordReset.UserID, &passwordReset.TokenHash,
			&passwordReset.ExpiresAt, &passwordReset.UsedAt, &passwordReset.UsedIP,
			&passwordReset.UserAgent, &passwordReset.CreatedAt, &passwordReset.UpdatedAt)

		if err != nil {
			r.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to scan password reset row, caused by %v", err)
			return nil, err
		}

		passwordResets = append(passwordResets, passwordReset)
	}

	if err := rows.Err(); err != nil {
		r.logger.Errorfctx(provider.AppLog, ctx, false, "Row iteration error, caused by %v", err)
		return nil, err
	}

	return passwordResets, nil
}

func (r *passwordResetRepo) RevokeUserPasswordResets(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()

	result, err := r.conn.Exec(ctx, revokeUserPasswordResetsQuery, userID, now)
	if err != nil {
		r.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to revoke user password resets, caused by %v", err)
		return err
	}

	rowsAffected := result.RowsAffected()
	r.logger.Infofctx(provider.AppLog, ctx, "User password resets revoked, user_id %s, resets_revoked %d", userID, rowsAffected)
	return nil
}
