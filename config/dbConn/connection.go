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

	// Step 1️⃣: Connect to default postgres DB
	defaultDSN := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=postgres port=%s sslmode=%s TimeZone=%s",
		config.Cfg.DbHost,
		config.Cfg.DbUser,
		config.Cfg.DbPassword,
		config.Cfg.DbPort,
		config.Cfg.DbSSLmode,
		config.Cfg.DbTimeZone,
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
		config.Cfg.DbHost,
		config.Cfg.DbUser,
		config.Cfg.DbPassword,
		dbName,
		config.Cfg.DbPort,
		config.Cfg.DbSSLmode,
		config.Cfg.DbTimeZone,
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
	)
	if err != nil {
		log.Fatalf("❌ Auto migration failed: %v", err)
	}

	log.Println("✅ Auto migration completed successfully with prefix 'mys_'")

	DB = db
	return DB
}
