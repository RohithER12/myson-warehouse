package repo

import (
	"context"
	dbconn "warehouse/config/dbConn"
	"warehouse/models"
)

type UserRepo struct{}

func NewUserRepo() *UserRepo { return &UserRepo{} }

func (r *UserRepo) Create(ctx context.Context, u *models.User) error {
	return dbconn.DB.WithContext(ctx).Create(u).Error
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	if err := dbconn.DB.WithContext(ctx).Where("email = ?", email).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}
