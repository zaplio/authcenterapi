package repository

import (
	"authcenterapi/internal/provider"
	model "authcenterapi/model/entity"
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type profileRepo struct {
	logger provider.ILogger
	conn   *pgx.Conn
}

// NewProfileRepository creates a new profile repository
func NewProfileRepository(logger provider.ILogger, conn *pgx.Conn) ProfileRepository {
	return &profileRepo{
		logger: logger,
		conn:   conn,
	}
}

const (
	insertProfileQuery = `
		INSERT INTO authentication.profiles 
		(id, user_id, first_name, last_name, display_name, date_of_birth, gender, avatar_url, 
		 bio, country, city, timezone, language, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

	getProfileByUserIDQuery = `
		SELECT id, user_id, first_name, last_name, display_name, date_of_birth, gender, 
			   avatar_url, bio, country, city, timezone, language, metadata, created_at, updated_at
		FROM authentication.profiles 
		WHERE user_id = $1`

	getUserWithProfileQuery = `
		SELECT u.id, u.username, u.email, u.email_verified_at, u.phone, u.phone_verified_at, 
			   u.status, u.last_login_at, u.last_login_ip, u.failed_login_attempts, u.locked_until, 
			   u.two_factor_enabled, u.two_factor_secret, u.created_at, u.updated_at, u.deleted_at,
			   p.id, p.first_name, p.last_name, p.display_name, p.date_of_birth, p.gender, 
			   p.avatar_url, p.bio, p.country, p.city, p.timezone, p.language, p.metadata, 
			   p.created_at, p.updated_at
		FROM authentication.users u
		LEFT JOIN authentication.profiles p ON u.id = p.user_id
		WHERE u.id = $1 AND u.deleted_at IS NULL`

	updateProfileQuery = `
		UPDATE authentication.profiles 
		SET first_name = $3, last_name = $4, display_name = $5, date_of_birth = $6, gender = $7,
			avatar_url = $8, bio = $9, country = $10, city = $11, timezone = $12, 
			language = $13, metadata = $14, updated_at = $15
		WHERE id = $1 AND user_id = $2`
)

func (r *profileRepo) Create(ctx context.Context, profile *model.Profile) error {
	if profile.ID == uuid.Nil {
		profile.ID = uuid.New()
	}

	now := time.Now()
	profile.CreatedAt = now
	profile.UpdatedAt = now

	_, err := r.conn.Exec(ctx, insertProfileQuery,
		profile.ID, profile.UserID, profile.FirstName, profile.LastName,
		profile.DisplayName, profile.DateOfBirth, profile.Gender, profile.AvatarURL,
		profile.Bio, profile.Country, profile.City, profile.Timezone,
		profile.Language, profile.Metadata, profile.CreatedAt, profile.UpdatedAt)

	if err != nil {
		r.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to create profile for user_id %s, caused by %v", profile.UserID, err)
		return err
	}

	r.logger.Infofctx(provider.AppLog, ctx, "Profile created successfully profile_id %s, user_id %s", profile.ID, profile.UserID)
	return nil
}

func (r *profileRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*model.Profile, error) {
	profile := &model.Profile{}

	err := r.conn.QueryRow(ctx, getProfileByUserIDQuery, userID).Scan(
		&profile.ID, &profile.UserID, &profile.FirstName, &profile.LastName,
		&profile.DisplayName, &profile.DateOfBirth, &profile.Gender, &profile.AvatarURL,
		&profile.Bio, &profile.Country, &profile.City, &profile.Timezone,
		&profile.Language, &profile.Metadata, &profile.CreatedAt, &profile.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to get profile by user ID %s, caused by %v", userID, err)
		return nil, err
	}

	return profile, nil
}

func (r *profileRepo) Update(ctx context.Context, profile *model.Profile) error {
	profile.UpdatedAt = time.Now()

	result, err := r.conn.Exec(ctx, updateProfileQuery,
		profile.ID, profile.UserID, profile.FirstName, profile.LastName,
		profile.DisplayName, profile.DateOfBirth, profile.Gender, profile.AvatarURL,
		profile.Bio, profile.Country, profile.City, profile.Timezone,
		profile.Language, profile.Metadata, profile.UpdatedAt)

	if err != nil {
		r.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to update profile, caused by %v", err)
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Infof(provider.AppLog, "No profile rows updated - profile_id: %s, user_id: %s", profile.ID, profile.UserID)
		return sql.ErrNoRows
	}

	r.logger.Infofctx(provider.AppLog, ctx, "Profile updated successfully profile_id %s, user_id %s", profile.ID, profile.UserID)
	return nil
}

func (r *profileRepo) GetUserWithProfile(ctx context.Context, userID uuid.UUID) (*model.UserWithProfile, error) {
	userWithProfile := &model.UserWithProfile{}
	user := &userWithProfile.User

	var profileID sql.NullString
	var firstName, lastName, displayName, avatarURL, bio, country, city sql.NullString
	var birthDate sql.NullTime
	var gender sql.NullString
	var timezone, language sql.NullString
	var metadata sql.NullString
	var profileCreatedAt, profileUpdatedAt sql.NullTime

	err := r.conn.QueryRow(ctx, getUserWithProfileQuery, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.EmailVerifiedAt, &user.Phone,
		&user.PhoneVerifiedAt, &user.Status, &user.LastLoginAt, &user.LastLoginIP,
		&user.FailedLoginAttempts, &user.LockedUntil, &user.TwoFactorEnabled,
		&user.TwoFactorSecret, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
		&profileID, &firstName, &lastName, &displayName, &birthDate, &gender,
		&avatarURL, &bio, &country, &city, &timezone, &language, &metadata,
		&profileCreatedAt, &profileUpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to get user with profile for user_id %s, caused by %v", userID, err)
		return nil, err
	}

	// Set profile if it exists
	if profileID.Valid {
		profile := &model.Profile{
			UserID:    userID,
			Timezone:  "UTC",
			Language:  "en",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if parsedID, err := uuid.Parse(profileID.String); err == nil {
			profile.ID = parsedID
		}

		if firstName.Valid {
			profile.FirstName = &firstName.String
		}
		if lastName.Valid {
			profile.LastName = &lastName.String
		}
		if displayName.Valid {
			profile.DisplayName = &displayName.String
		}
		if birthDate.Valid {
			profile.DateOfBirth = &birthDate.Time
		}
		if gender.Valid {
			profile.Gender = &gender.String
		}
		if avatarURL.Valid {
			profile.AvatarURL = &avatarURL.String
		}
		if bio.Valid {
			profile.Bio = &bio.String
		}
		if country.Valid {
			profile.Country = &country.String
		}
		if city.Valid {
			profile.City = &city.String
		}
		if timezone.Valid {
			profile.Timezone = timezone.String
		}
		if language.Valid {
			profile.Language = language.String
		}
		if metadata.Valid {
			if err := json.Unmarshal([]byte(metadata.String), &profile.Metadata); err != nil {
				r.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to unmarshal metadata, caused by %v", err)
			}
		}
		if profileCreatedAt.Valid {
			profile.CreatedAt = profileCreatedAt.Time
		}
		if profileUpdatedAt.Valid {
			profile.UpdatedAt = profileUpdatedAt.Time
		}

		userWithProfile.Profile = profile
	}

	return userWithProfile, nil
}
