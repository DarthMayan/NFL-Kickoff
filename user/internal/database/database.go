package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"kickoff.com/user/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Connect establece la conexión con la base de datos PostgreSQL
func Connect() error {
	// Obtener configuración desde variables de entorno
	host := getEnv("DB_HOST", "postgres-service")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "kickoff_user")
	password := getEnv("DB_PASSWORD", "kickoff_password_123")
	dbname := getEnv("DB_NAME", "user_db")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})

	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	// Configurar pool de conexiones
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("✅ Connected to PostgreSQL database:", dbname)

	// Auto-migrar modelos
	if err := AutoMigrate(); err != nil {
		return fmt.Errorf("failed to auto-migrate: %w", err)
	}

	return nil
}

// AutoMigrate ejecuta las migraciones automáticas
func AutoMigrate() error {
	log.Println("Running auto-migration for User service...")
	return DB.AutoMigrate(
		&models.User{},
	)
}

// Close cierra la conexión a la base de datos
func Close() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// getEnv obtiene una variable de entorno o devuelve un valor por defecto
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
