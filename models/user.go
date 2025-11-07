package models

import (
	"time"

	"gorm.io/gorm"
)

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleEmployee Role = "employee"
)

type User struct {
	ID           uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	WarehouseID  uint           `gorm:"not null;index" json:"warehouse_id"`
	Warehouse    Warehouse      `gorm:"foreignKey:WarehouseID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"warehouse"`
	Name         string         `gorm:"type:varchar(255);not null" json:"name"`
	Email        string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"`
	Role         Role           `gorm:"type:varchar(20);not null;default:'employee'" json:"role"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type RegisterDTO struct {
	Name            string `json:"name" binding:"required"`
	WarehouseID     uint   `json:"warehouse_id" binding:"required"`
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required,min=8"`
	// Role     Role   `json:"role" binding:"required,oneof=admin employee"`
}

type LoginDTO struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
