package repo

import (
	"context"
	"fmt"
	"log"
	dbconn "warehouse/config/dbConn"
	"warehouse/models"
)

type UserRepo struct{}

func NewUserRepo() *UserRepo { return &UserRepo{} }

// Create inserts a new user into the database
func (r *UserRepo) Create(ctx context.Context, u *models.User) error {
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy
	table := ns.TableName("User")

	if err := db.Table(table).Create(u).Error; err != nil {
		return fmt.Errorf("failed to create user (email: %s): %w", u.Email, err)
	}

	log.Printf("👤 New user created: ID=%d, Role=%s, Email=%s", u.ID, u.Role, u.Email)
	return nil
}

// GetByEmail retrieves a user by their email address
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy
	table := ns.TableName("User")

	var user models.User
	if err := db.Table(table).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to find user by email %s: %w", email, err)
	}

	return &user, nil
}
