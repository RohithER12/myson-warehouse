package dbconn

import (
	"fmt"
	"log"
	"warehouse/config"
	"warehouse/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// postgres instance
var DB *gorm.DB

func ConnectDB() *gorm.DB {
	config.LoadConfig()
	dbName := config.Cfg.DbName
	dbHost := config.Cfg.DbHost
	dbUser := config.Cfg.DbUser
	dbPassword := config.Cfg.DbPassword
	dbPort := config.Cfg.DbPort
	dbSSLmode := config.Cfg.DbSSLmode
	dbTimeZone := config.Cfg.DbTimeZone

	// Step 1️⃣: Connect to default postgres DB
	defaultDSN := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=postgres port=%s sslmode=%s TimeZone=%s",
		dbHost,
		dbUser,
		dbPassword,
		dbPort,
		dbSSLmode,
		dbTimeZone,
	)

	defaultDB, err := gorm.Open(postgres.Open(defaultDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to PostgreSQL server: %v", err)
	}

	// Step 2️⃣: Check if target DB exists
	var exists bool
	checkQuery := fmt.Sprintf("SELECT EXISTS (SELECT datname FROM pg_database WHERE datname = '%s')", dbName)
	if err := defaultDB.Raw(checkQuery).Scan(&exists).Error; err != nil {
		log.Fatalf("❌ Failed to check database existence: %v", err)
	}

	// Step 3️⃣: Create DB if missing
	if !exists {
		createQuery := fmt.Sprintf("CREATE DATABASE %s;", dbName)
		log.Printf("📦 Creating new database: %s", dbName)
		if err := defaultDB.Exec(createQuery).Error; err != nil {
			log.Fatalf("❌ Failed to create database: %v", err)
		}
	} else {
		log.Printf("✅ Database '%s' already exists", dbName)
	}

	// Step 4️⃣: Connect to actual target DB
	targetDSN := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		dbHost,
		dbUser,
		dbPassword,
		dbName,
		dbPort,
		dbSSLmode,
		dbTimeZone,
	)

	// ✅ Add NamingStrategy to prefix all tables with `mys_`
	db, err := gorm.Open(postgres.Open(targetDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "mys_", // ✅ Add prefix to all table names
			SingularTable: true,   // optional: prevent pluralization (mys_user instead of mys_users)
		},
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to target database: %v", err)
	}

	// Step 5️⃣: Run migrations
	err = db.AutoMigrate(
		&models.Warehouse{},
		&models.RentRate{},
		&models.User{},
		&models.Supplier{},
		&models.Product{},
		&models.Profit{},
		&models.Billing{},
		&models.BillingItem{},
		&models.Batch{},
		&models.BatchProductEntry{},
		&models.OnBoardExpense{},
		&models.OffBoardExpense{},
	)
	if err != nil {
		log.Fatalf("❌ Auto migration failed: %v", err)
	}

	log.Println("✅ Auto migration completed successfully with prefix 'mys_'")

	// Step 6️⃣: Create index optimizations
	if err := CreateRequiredIndexes(db); err != nil {
		log.Fatalf("❌ Failed creating indexes: %v", err)
	}

	DB = db
	return DB
}

// CreateRequiredIndexes ensures all performance-related indexes exist.
func CreateRequiredIndexes(db *gorm.DB) error {
	indexQueries := []string{
		// 📦 Batch
		`CREATE INDEX IF NOT EXISTS idx_batch_warehouse ON mys_batch (warehouse_id);`,
		`CREATE INDEX IF NOT EXISTS idx_batch_created_at ON mys_batch (created_at);`,

		// 📦 BatchProductEntry
		`CREATE INDEX IF NOT EXISTS idx_bpe_batch ON mys_batch_product_entry (batch_id);`,
		`CREATE INDEX IF NOT EXISTS idx_bpe_product ON mys_batch_product_entry (product_id);`,
		`CREATE INDEX IF NOT EXISTS idx_bpe_created_at ON mys_batch_product_entry (created_at);`,

		// 🧾 BillingItem
		`CREATE INDEX IF NOT EXISTS idx_bi_batch ON mys_billing_item (batch_id);`,
		`CREATE INDEX IF NOT EXISTS idx_bi_product ON mys_billing_item (product_id);`,
		`CREATE INDEX IF NOT EXISTS idx_bi_created_at ON mys_billing_item (created_at);`,

		// 💰 Profit
		`CREATE INDEX IF NOT EXISTS idx_profit_batch ON mys_profit (batch_id);`,
		`CREATE INDEX IF NOT EXISTS idx_profit_product ON mys_profit (product_id);`,

		// 📦 Warehouse
		`CREATE INDEX IF NOT EXISTS idx_warehouse_rentconfig ON mys_warehouse (rent_config_id);`,

		// 🛒 Product
		`CREATE INDEX IF NOT EXISTS idx_product_supplier ON mys_product (supplier_id);`,
		`CREATE INDEX IF NOT EXISTS idx_product_category ON mys_product (category);`,
	}

	for _, q := range indexQueries {
		if err := db.Exec(q).Error; err != nil {
			return fmt.Errorf("failed to create index: %v\nQuery: %s", err, q)
		}
	}

	log.Println("✅ Required indexes created successfully")
	return nil
}
