package repo

import (
	"context"
	"fmt"
	"log"
	dbconn "warehouse/config/dbConn"
	"warehouse/models"
)

type ProductRepo struct {
}

// NewProductRepo initializes the repository
func NewProductRepo() *ProductRepo {
	return &ProductRepo{}
}

// Create inserts a new product
func (r *ProductRepo) Create(ctx context.Context, product *models.Product) (uint, error) {
	// ✅ Validate supplier ID before inserting
	if product.SupplierID == 0 {
		return 0, fmt.Errorf("invalid supplier_id: must reference an existing supplier")
	}

	// ✅ Check if supplier exists
	var exists bool
	if err := dbconn.DB.WithContext(ctx).
		Model(&models.Supplier{}).
		Select("count(*) > 0").
		Where("id = ?", product.SupplierID).
		Find(&exists).Error; err != nil {
		return 0, fmt.Errorf("failed to verify supplier: %w", err)
	}
	if !exists {
		return 0, fmt.Errorf("supplier with id %d not found", product.SupplierID)
	}

	// ✅ Insert product
	if err := dbconn.DB.WithContext(ctx).Create(product).Error; err != nil {
		return 0, err
	}

	log.Printf("🛠 New product created: %d", product.ID)
	return product.ID, nil
}

// GetByID fetches a product by ID
func (r *ProductRepo) GetByID(ctx context.Context, id uint) (*models.Product, error) {
	var product models.Product
	if err := dbconn.DB.WithContext(ctx).Preload("Supplier").
		Find(&product).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

// GetAll fetches all products
func (r *ProductRepo) GetAll(ctx context.Context) ([]models.Product, error) {
	var products []models.Product
	if err := dbconn.DB.WithContext(ctx).
		Preload("Supplier").
		Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

// Update modifies a product
func (r *ProductRepo) Update(ctx context.Context, id uint, update map[string]interface{}) error {
	if err := dbconn.DB.WithContext(ctx).
		Model(&models.Product{}).
		Where("id = ?", id).
		Updates(update).Error; err != nil {
		return err
	}
	return nil
}

// Delete removes a product
func (r *ProductRepo) Delete(ctx context.Context, id uint) error {
	if err := dbconn.DB.WithContext(ctx).
		Delete(&models.Product{}, id).Error; err != nil {
		return err
	}
	return nil
}

func (r *ProductRepo) GetAllProductCategories(ctx context.Context) ([]string, error) {
	db := dbconn.DB.WithContext(ctx)
	var categories []string

	err := db.
		Table("products").
		Select("DISTINCT category").
		Where("category IS NOT NULL AND category <> ''").
		Order("category ASC").
		Pluck("category", &categories).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch product categories: %v", err)
	}

	return categories, nil
}
