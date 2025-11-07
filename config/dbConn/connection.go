package dbconn

import (
	"fmt"
	"log"
	"warehouse/config"
	"warehouse/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

	db, err := gorm.Open(postgres.Open(targetDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to target database: %v", err)
	}

	// Step 5️⃣: Run migrations
	err = db.AutoMigrate(
		&models.Batch{},
		&models.BatchProductEntry{},
		&models.Warehouse{},
		&models.RentRate{},
		&models.Product{},
		&models.Supplier{},
		&models.Billing{},
		&models.BillingItem{},
		&models.Profit{},
		&models.User{},
	)
	if err != nil {
		log.Fatalf("❌ Auto migration failed: %v", err)
	}

	log.Println("✅ Auto migration completed successfully")
	DB = db
	return DB
}
