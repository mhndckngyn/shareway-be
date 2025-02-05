package repository

import (
	"shareway/infra/db/migration"
	"shareway/schemas"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type AdminRepository struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewAdminRepository(db *gorm.DB, redis *redis.Client) IAdminRepository {
	return &AdminRepository{
		db:    db,
		redis: redis,
	}
}

type IAdminRepository interface {
	CheckAdminExists(req schemas.LoginAdminRequest) (migration.Admin, error)
	GetAdminProfile(adminID uuid.UUID) (migration.Admin, error)
}

// CheckAdminExists checks if the admin exists in the database
func (r *AdminRepository) CheckAdminExists(req schemas.LoginAdminRequest) (migration.Admin, error) {
	var admin migration.Admin
	if err := r.db.Where("username = ?", req.Username).First(&admin).Error; err != nil {
		return admin, err
	}
	return admin, nil
}

// GetAdminProfile gets the profile of the admin
func (r *AdminRepository) GetAdminProfile(adminID uuid.UUID) (migration.Admin, error) {
	var admin migration.Admin
	if err := r.db.Where("id = ?", adminID).First(&admin).Error; err != nil {
		return admin, err
	}
	return admin, nil
}
