package repository

import (
	"context"
	"gin/user-management-api/internal/db/sqlc"

	"github.com/google/uuid"
)

type UserRepository interface {
	GetAll(ctx context.Context, search, orderBy, sort string, limit, offset int32) ([]sqlc.User, error)
	GetAllV2(ctx context.Context, search string, orderBy, sort string, limit, offset int32, deleted bool) ([]sqlc.User, error)
	CountUsers(ctx context.Context, search string, deleted bool) (int64, error)
	Create(ctx context.Context, userParams sqlc.CreateUserParams) (sqlc.User, error)
	FindByUUID(ctx context.Context, userUuid uuid.UUID) (sqlc.User, error)
	Update(ctx context.Context, userParams sqlc.UpdateUserByUuidParams) (sqlc.User, error)
	SoftDelete(ctx context.Context, userUuid uuid.UUID) (sqlc.User, error)
	Restore(ctx context.Context, userUuid uuid.UUID) (sqlc.User, error)
	Delete(ctx context.Context, userUuid uuid.UUID) (sqlc.User, error)
	GetByEmail(ctx context.Context, email string) (sqlc.User, error)
	UpdatePassword(ctx context.Context, input sqlc.UpdatePasswordParams) (sqlc.User, error)
}
