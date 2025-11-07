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
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy
	productTable := ns.TableName("Product")
	supplierTable := ns.TableName("Supplier")

	// ✅ Validate supplier ID
	if product.SupplierID == 0 {
		return 0, fmt.Errorf("invalid supplier_id: must reference an existing supplier")
	}

	// ✅ Check if supplier exists
	var exists bool
	if err := db.Table(supplierTable).
		Select("count(*) > 0").
		Where("id = ?", product.SupplierID).
		Find(&exists).Error; err != nil {
		return 0, fmt.Errorf("failed to verify supplier existence: %w", err)
	}
	if !exists {
		return 0, fmt.Errorf("supplier with id %d not found", product.SupplierID)
	}

	// ✅ Insert product
	if err := db.Table(productTable).Create(product).Error; err != nil {
		return 0, fmt.Errorf("failed to create product: %w", err)
	}

	log.Printf("🧩 New product created: ID=%d, Name=%s, SupplierID=%d", product.ID, product.Name, product.SupplierID)
	return product.ID, nil
}

// GetByID fetches a product by ID
func (r *ProductRepo) GetByID(ctx context.Context, id uint) (*models.Product, error) {
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy
	table := ns.TableName("Product")

	var product models.Product
	if err := db.Table(table).
		Preload("Supplier").
		First(&product, id).Error; err != nil {
		return nil, fmt.Errorf("failed to find product with ID %d: %w", id, err)
	}

	return &product, nil
}

// GetAll fetches all products
func (r *ProductRepo) GetAll(ctx context.Context) ([]models.Product, error) {
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy
	table := ns.TableName("Product")

	var products []models.Product
	if err := db.Table(table).
		Preload("Supplier").
		Find(&products).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch products: %w", err)
	}

	log.Printf("📦 Retrieved %d products", len(products))
	return products, nil
}

// Update modifies a product
func (r *ProductRepo) Update(ctx context.Context, id uint, update map[string]interface{}) error {
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy
	table := ns.TableName("Product")

	if err := db.Table(table).
		Where("id = ?", id).
		Updates(update).Error; err != nil {
		return fmt.Errorf("failed to update product ID %d: %w", id, err)
	}

	log.Printf("📝 Product updated: ID=%d", id)
	return nil
}

// Delete removes a product
func (r *ProductRepo) Delete(ctx context.Context, id uint) error {
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy
	table := ns.TableName("Product")

	if err := db.Table(table).
		Delete(&models.Product{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete product ID %d: %w", id, err)
	}

	log.Printf("🗑️ Product deleted: ID=%d", id)
	return nil
}

// GetAllProductCategories fetches all distinct product categories
func (r *ProductRepo) GetAllProductCategories(ctx context.Context) ([]string, error) {
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy
	table := ns.TableName("Product")

	var categories []string
	if err := db.Table(table).
		Select("DISTINCT category").
		Where("category IS NOT NULL AND category <> ''").
		Order("category ASC").
		Pluck("category", &categories).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch product categories: %w", err)
	}

	log.Printf("📊 Retrieved %d distinct categories", len(categories))
	return categories, nil
}
