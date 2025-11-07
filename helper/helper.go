package helper

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
	dbconn "warehouse/config/dbConn"
	"warehouse/models"
	"warehouse/repo"

	"golang.org/x/crypto/bcrypt"
)

func GetDurationRange(duration string) (time.Time, time.Time) {
	now := time.Now()
	var start time.Time

	switch duration {
	case "lastweek":
		start = now.AddDate(0, 0, -7)
	case "lastmonth":
		start = now.AddDate(0, -1, 0)
	case "lastyear":
		start = now.AddDate(-1, 0, 0)
	default:
		start = now.AddDate(0, 0, -30)
	}

	return start, now
}

func GetOrCreateProductData(m map[uint]*models.ProductWiseData, productID uint) *models.ProductWiseData {
	if m[productID] == nil {
		m[productID] = &models.ProductWiseData{}
	}
	return m[productID]
}

// StartHealthPing periodically checks both the HTTP server and the PostgreSQL database.
func StartHealthPing(baseURL string, interval time.Duration) {
	log.Println("✅ Server and the PostgreSQL database Health check is Running..............🚀.")
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {

		resp, err := http.Get(fmt.Sprintf("%s/ping", baseURL))
		if err != nil {
			log.Printf("❌ Server health check failed: %v", err)
		} else {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				log.Printf("⚠️ Server returned non-OK: %d - %s", resp.StatusCode, string(body))
			} else {
				log.Printf("✅ Server OK [%s] at %s", baseURL, time.Now().Format(time.RFC3339))
			}
		}

		sqlDB, err := dbconn.DB.DB()
		if err != nil {
			log.Printf("❌ Could not access DB instance: %v", err)
			continue
		}

		if err := sqlDB.Ping(); err != nil {
			log.Printf("❌ Database health check failed: %v", err)
		} else {
			log.Printf("✅ Database OK at %s", time.Now().Format(time.RFC3339))
		}
	}
}

func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(b), err
}

func CheckPassword(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}

func EnsureAdmin() {

	ur := repo.NewUserRepo()
	const adminEmail = "admin@myson.com"
	ctx := context.Background()
	if _, err := ur.GetByEmail(ctx, adminEmail); err == nil {
		return
	}
	hash, _ := HashPassword("Secretpassword@123")

	if err := ur.Create(ctx, &models.User{
		Name:         "Super Admin",
		Email:        adminEmail,
		PasswordHash: hash,
		Role:         models.RoleAdmin,
		WarehouseID:  1,
	}); err != nil {
		log.Printf("failed to seed admin: %v", err)
	}
}
