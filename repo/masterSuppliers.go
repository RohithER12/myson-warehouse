package repo

import (
	"context"
	"log"
	dbconn "warehouse/config/dbConn"
	"warehouse/models"
)

type SupplierRepo struct {
}

// NewSupplierRepo initializes the repository
func NewSupplierRepo() *SupplierRepo {
	return &SupplierRepo{}
}

// Create inserts a new supplier
func (r *SupplierRepo) Create(ctx context.Context, supplier *models.Supplier) (uint, error) {

	if err := dbconn.DB.WithContext(ctx).Create(&supplier).Error; err != nil {
		return 0, err
	}

	log.Printf("🛠 New supplier created: %d", supplier.ID)
	return supplier.ID, nil
}

// GetByID fetches a supplier by ID
func (r *SupplierRepo) GetByID(ctx context.Context, id uint) (*models.Supplier, error) {

	var supplier models.Supplier
	if err := dbconn.DB.WithContext(ctx).First(&supplier, id).Error; err != nil {
		return nil, err
	}
	return &supplier, nil
}

// GetAll fetches all suppliers
func (r *SupplierRepo) GetAll(ctx context.Context) ([]models.Supplier, error) {

	var suppliers []models.Supplier
	if err := dbconn.DB.WithContext(ctx).Find(&suppliers).Error; err != nil {
		return nil, err
	}
	return suppliers, nil
}

// Update modifies a supplier
func (r *SupplierRepo) Update(ctx context.Context, update models.Supplier) error {

	if err := dbconn.DB.WithContext(ctx).
		Model(&models.Supplier{}).
		Where("id = ?", update.ID).
		Updates(update).Error; err != nil {
		return err
	}
	return nil
}

// Delete removes a supplier
func (r *SupplierRepo) Delete(ctx context.Context, id uint) error {
	if err := dbconn.DB.WithContext(ctx).
		Delete(&models.Supplier{}, id).Error; err != nil {
		return err
	}
	return nil
}
