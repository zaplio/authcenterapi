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

type userRepo struct {
	logger provider.ILogger
	conn   *pgx.Conn
}

// NewUserRepository creates a new user repository
func NewUserRepository(logger provider.ILogger, conn *pgx.Conn) UserRepository {
	return &userRepo{
		logger: logger,
		conn:   conn,
	}
}

const (
	insertUserQuery = `
		INSERT INTO authentication.users 
		(id, username, email, password_hash, phone, status, two_factor_enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	getUserByIDQuery = `
		SELECT id, username, email, email_verified_at, password_hash, phone, phone_verified_at, 
			   status, last_login_at, last_login_ip, failed_login_attempts, locked_until, 
			   two_factor_enabled, two_factor_secret, created_at, updated_at, deleted_at
		FROM authentication.users 
		WHERE id = $1 AND deleted_at IS NULL`

	getUserByEmailQuery = `
		SELECT id, username, email, email_verified_at, password_hash, phone, phone_verified_at, 
			   status, last_login_at, last_login_ip, failed_login_attempts, locked_until, 
			   two_factor_enabled, two_factor_secret, created_at, updated_at, deleted_at
		FROM authentication.users 
		WHERE email = $1 AND deleted_at IS NULL`

	getUserByUsernameOrEmailQuery = `
		SELECT id, username, email, email_verified_at, password_hash, phone, phone_verified_at, 
			   status, last_login_at, last_login_ip, failed_login_attempts, locked_until, 
			   two_factor_enabled, two_factor_secret, created_at, updated_at, deleted_at
		FROM authentication.users 
		WHERE (username = $1 OR email = $1) AND deleted_at IS NULL`

	updateUserQuery = `
		UPDATE authentication.users 
		SET username = $2, email = $3, phone = $4, status = $5, email_verified_at = $6, 
			phone_verified_at = $7, two_factor_enabled = $8, two_factor_secret = $9, updated_at = $10
		WHERE id = $1`

	updateLoginInfoQuery = `
		UPDATE authentication.users 
		SET last_login_at = $2, last_login_ip = $3, failed_login_attempts = 0, updated_at = $4
		WHERE id = $1`

	incrementFailedAttemptsQuery = `
		UPDATE authentication.users 
		SET failed_login_attempts = failed_login_attempts + 1,
			locked_until = CASE 
				WHEN failed_login_attempts + 1 >= 5 THEN $2
				ELSE locked_until
			END,
			updated_at = $3
		WHERE id = $1`

	resetFailedAttemptsQuery = `
		UPDATE authentication.users 
		SET failed_login_attempts = 0, locked_until = NULL, updated_at = $2
		WHERE id = $1`

	updatePasswordQuery = `
		UPDATE authentication.users 
		SET password_hash = $2, updated_at = $3
		WHERE id = $1`
)

func (r *userRepo) Create(ctx context.Context, user *model.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := r.conn.Exec(ctx, insertUserQuery,
		user.ID, user.Username, user.Email, user.PasswordHash, user.Phone,
		user.Status, user.TwoFactorEnabled, user.CreatedAt, user.UpdatedAt)

	if err != nil {
		r.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to create user for username %s, caused by %v", user.Username, err)
		return err
	}

	r.logger.Infofctx(provider.AppLog, ctx, "User created successfully user_id %s, username %s", user.ID, user.Username)
	return nil
}

func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user := &model.User{}

	err := r.conn.QueryRow(ctx, getUserByIDQuery, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.EmailVerifiedAt, &user.PasswordHash,
		&user.Phone, &user.PhoneVerifiedAt, &user.Status, &user.LastLoginAt, &user.LastLoginIP,
		&user.FailedLoginAttempts, &user.LockedUntil, &user.TwoFactorEnabled, &user.TwoFactorSecret,
		&user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to get user by user_id %s, caused by %v", id, err)
		return nil, err
	}

	return user, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &model.User{}

	err := r.conn.QueryRow(ctx, getUserByEmailQuery, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.EmailVerifiedAt, &user.PasswordHash,
		&user.Phone, &user.PhoneVerifiedAt, &user.Status, &user.LastLoginAt, &user.LastLoginIP,
		&user.FailedLoginAttempts, &user.LockedUntil, &user.TwoFactorEnabled, &user.TwoFactorSecret,
		&user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to get user by email %s, caused by %v", email, err)
		return nil, err
	}

	return user, nil
}

func (r *userRepo) GetByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (*model.User, error) {
	user := &model.User{}

	err := r.conn.QueryRow(ctx, getUserByUsernameOrEmailQuery, usernameOrEmail).Scan(
		&user.ID, &user.Username, &user.Email, &user.EmailVerifiedAt, &user.PasswordHash,
		&user.Phone, &user.PhoneVerifiedAt, &user.Status, &user.LastLoginAt, &user.LastLoginIP,
		&user.FailedLoginAttempts, &user.LockedUntil, &user.TwoFactorEnabled, &user.TwoFactorSecret,
		&user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to get user by username or email %s, caused by %v", usernameOrEmail, err)
		return nil, err
	}

	return user, nil
}

func (r *userRepo) Update(ctx context.Context, user *model.User) error {
	user.UpdatedAt = time.Now()

	_, err := r.conn.Exec(ctx, updateUserQuery,
		user.ID, user.Username, user.Email, user.Phone, user.Status,
		user.EmailVerifiedAt, user.PhoneVerifiedAt, user.TwoFactorEnabled,
		user.TwoFactorSecret, user.UpdatedAt)

	if err != nil {
		r.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to update user for user_id %s, caused by %v", user.ID, err)
		return err
	}

	r.logger.Infofctx(provider.AppLog, ctx, "User updated successfully user_id %s", user.ID)
	return nil
}

func (r *userRepo) UpdateLoginInfo(ctx context.Context, userID uuid.UUID, loginIP net.IP) error {
	now := time.Now()

	_, err := r.conn.Exec(ctx, updateLoginInfoQuery, userID, now, loginIP, now)
	if err != nil {
		r.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to update login info for user_id %s, caused by %v", userID, err)
		return err
	}

	r.logger.Infofctx(provider.AppLog, ctx, "Login info updated for user_id %s, login_ip %s", userID, loginIP)
	return nil
}

func (r *userRepo) IncrementFailedAttempts(ctx context.Context, userID uuid.UUID) error {
	lockUntil := time.Now().Add(30 * time.Minute)
	now := time.Now()

	_, err := r.conn.Exec(ctx, incrementFailedAttemptsQuery, userID, lockUntil, now)
	if err != nil {
		r.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to increment failed attempts for user_id %s, caused by %v", userID, err)
		return err
	}

	r.logger.Infofctx(provider.AppLog, ctx, "Failed login attempts incremented for user_id %s", userID)
	return nil
}

func (r *userRepo) ResetFailedAttempts(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()

	_, err := r.conn.Exec(ctx, resetFailedAttemptsQuery, userID, now)
	if err != nil {
		r.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to reset failed attempts for user_id %s, caused by %v", userID, err)
		return err
	}

	r.logger.Infofctx(provider.AppLog, ctx, "Failed login attempts reset for user_id %s", userID)
	return nil
}

func (r *userRepo) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	now := time.Now()

	_, err := r.conn.Exec(ctx, updatePasswordQuery, userID, passwordHash, now)
	if err != nil {
		r.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to update password for user_id %s, caused by %v", userID, err)
		return err
	}

	r.logger.Infofctx(provider.AppLog, ctx, "Password updated successfully for user_id %s", userID)
	return nil
}
