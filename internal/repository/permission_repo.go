package repository

import (
	"authcenterapi/internal/provider"
	model "authcenterapi/model/entity"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type permissionRepo struct {
	logger provider.ILogger
	conn   *pgx.Conn
}

// NewPermissionRepository creates a new permission repository
func NewPermissionRepository(logger provider.ILogger, conn *pgx.Conn) PermissionRepository {
	return &permissionRepo{
		logger: logger,
		conn:   conn,
	}
}

const (
	getUserPermissionsQuery = `
		SELECT id, user_id, role, permission, resource, resource_id, granted_by, granted_at, expires_at,
			   revoked, revoked_at, revoked_by, conditions, metadata, created_at, updated_at
		FROM authentication.permissions 
		WHERE user_id = $1 AND revoked = false
		ORDER BY created_at DESC`
)

func (r *permissionRepo) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*model.Permission, error) {
	rows, err := r.conn.Query(ctx, getUserPermissionsQuery, userID)
	if err != nil {
		r.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to get user permissions for user_id %s, caused by %v", userID, err)
		return nil, err
	}
	defer rows.Close()

	var permissions []*model.Permission
	for rows.Next() {
		permission := &model.Permission{}
		err := rows.Scan(
			&permission.ID, &permission.UserID, &permission.Role, &permission.Permission,
			&permission.Resource, &permission.ResourceID, &permission.GrantedBy, &permission.GrantedAt,
			&permission.ExpiresAt, &permission.Revoked, &permission.RevokedAt, &permission.RevokedBy,
			&permission.Conditions, &permission.Metadata, &permission.CreatedAt, &permission.UpdatedAt)

		if err != nil {
			r.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to scan permission row, caused by %v", err)
			return nil, err
		}

		permissions = append(permissions, permission)
	}

	if err := rows.Err(); err != nil {
		r.logger.Errorfctx(provider.AppLog, ctx, false, "Row iteration error, caused by %v", err)
		return nil, err
	}

	return permissions, nil
}
