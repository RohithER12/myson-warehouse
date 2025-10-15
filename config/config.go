package config

import (
	"fmt"
	"log"
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
	DBConnectionString string `mapstructure:"DB_CONNECTION_STRING"`
	DBName             string `mapstructure:"DB_NAME"`
	JWTSecret          string `mapstructure:"JWT_SECRET"`
}

var (
	Cfg  Config
	once sync.Once
)

func LoadConfig() {
	once.Do(
		func() {
			log.Println("config loading")
			viper.SetConfigFile(".env")
			viper.AddConfigPath(".")
			viper.AutomaticEnv()

			if err := viper.ReadInConfig(); err != nil {
				log.Fatalf("❌ Error reading config file: %v", err)
			}

			if err := viper.Unmarshal(&Cfg); err != nil {
				log.Fatalf("❌ Unable to decode config into struct: %v", err)
			}

			fmt.Println("✅ Config loaded successfully")
			// fmt.Printf("📦 DB: %s\n DBConnectionString : %s", Cfg.DBName, Cfg.DBConnectionString)
		},
	)

}
