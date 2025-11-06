package repo

import (
	"context"
	"fmt"
	"log"
	"time"
	dbconn "warehouse/config/dbConn"
	"warehouse/models"
)

type WarehouseRepo struct {
}

// NewWarehouseRepo initializes the repository
func NewWarehouseRepo() *WarehouseRepo {
	return &WarehouseRepo{}
}

// Create inserts a new warehouse
func (r *WarehouseRepo) Create(ctx context.Context, warehouse *models.Warehouse) (uint, error) {
	warehouse.CreatedAt = time.Now()
	warehouse.UpdatedAt = time.Now()

	// Save to PostgreSQL using GORM
	if err := dbconn.DB.WithContext(ctx).Create(&warehouse).Error; err != nil {
		return 0, err
	}

	log.Printf("🏠 New warehouse created: %d", warehouse.ID)
	return warehouse.ID, nil
}

// GetByID fetches a warehouse by ID
func (r *WarehouseRepo) GetByID(ctx context.Context, id uint) (*models.Warehouse, error) {
	var warehouse models.Warehouse
	if err := dbconn.DB.WithContext(ctx).Preload("RentConfig").
		Find(&warehouse).Error; err != nil {
		return nil, err
	}
	return &warehouse, nil
}

// GetAll fetches all warehouses
func (r *WarehouseRepo) GetAll(ctx context.Context) ([]models.Warehouse, error) {
	var warehouses []models.Warehouse
	if err := dbconn.DB.WithContext(ctx).
		Preload("RentConfig").
		Find(&warehouses).Error; err != nil {
		return nil, err
	}
	return warehouses, nil
}

// Update modifies a warehouse record
func (r *WarehouseRepo) Update(ctx context.Context, warehouse *models.Warehouse) error {
	var existing models.Warehouse

	// Fetch existing warehouse to get current RentConfig ID
	if err := dbconn.DB.WithContext(ctx).
		Preload("RentConfig").
		First(&existing, warehouse.ID).Error; err != nil {
		return fmt.Errorf("warehouse not found: %w", err)
	}

	// Update warehouse fields (excluding relations)
	if err := dbconn.DB.WithContext(ctx).
		Model(&models.Warehouse{}).
		Where("id = ?", warehouse.ID).
		Omit("RentConfig").
		Updates(warehouse).Error; err != nil {
		return err
	}

	// ✅ Always update related RentConfig using existing ID
	if existing.RentConfig.ID != 0 {
		warehouse.RentConfig.ID = existing.RentConfig.ID // ensure correct ID
		if err := dbconn.DB.WithContext(ctx).
			Model(&models.RentRate{}).
			Where("id = ?", existing.RentConfig.ID).
			Updates(warehouse.RentConfig).Error; err != nil {
			return err
		}
	}

	return nil
}

// Delete removes a warehouse
func (r *WarehouseRepo) Delete(ctx context.Context, id uint) error {
	if err := dbconn.DB.WithContext(ctx).
		Delete(&models.Warehouse{}, id).Error; err != nil {
		return err
	}
	return nil
}
