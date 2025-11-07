package repo

import (
	"context"
	"fmt"
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
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy
	table := ns.TableName("Supplier")

	if err := db.Table(table).Create(&supplier).Error; err != nil {
		return 0, fmt.Errorf("failed to create supplier: %w", err)
	}

	log.Printf("🛠 New supplier created: ID=%d, Name=%s", supplier.ID, supplier.Name)
	return supplier.ID, nil
}

// GetByID fetches a supplier by ID
func (r *SupplierRepo) GetByID(ctx context.Context, id uint) (*models.Supplier, error) {
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy
	table := ns.TableName("Supplier")

	var supplier models.Supplier
	if err := db.Table(table).First(&supplier, id).Error; err != nil {
		return nil, fmt.Errorf("failed to find supplier with ID %d: %w", id, err)
	}
	return &supplier, nil
}

// GetAll fetches all suppliers
func (r *SupplierRepo) GetAll(ctx context.Context) ([]models.Supplier, error) {
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy
	table := ns.TableName("Supplier")

	var suppliers []models.Supplier
	if err := db.Table(table).Find(&suppliers).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch suppliers: %w", err)
	}

	log.Printf("📦 Retrieved %d suppliers", len(suppliers))
	return suppliers, nil
}

// Update modifies a supplier
func (r *SupplierRepo) Update(ctx context.Context, update models.Supplier) error {
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy
	table := ns.TableName("Supplier")

	if err := db.Table(table).
		Where("id = ?", update.ID).
		Updates(update).Error; err != nil {
		return fmt.Errorf("failed to update supplier ID %d: %w", update.ID, err)
	}

	log.Printf("🔄 Supplier updated: ID=%d, Name=%s", update.ID, update.Name)
	return nil
}

// Delete removes a supplier
func (r *SupplierRepo) Delete(ctx context.Context, id uint) error {
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy
	table := ns.TableName("Supplier")

	if err := db.Table(table).Delete(&models.Supplier{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete supplier ID %d: %w", id, err)
	}

	log.Printf("🗑️ Supplier deleted: ID=%d", id)
	return nil
}
