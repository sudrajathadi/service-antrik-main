package config

import (
	"fmt"
	"os"

	"service-antrik-chatbot/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectDatabase() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta",
		getEnvAny([]string{"DB_HOST", "DB_HOST_GO"}, "localhost"),
		getEnvAny([]string{"DB_USER", "DB_USER_GO"}, "postgres"),
		getEnvAny([]string{"DB_PASSWORD", "DB_PASSWORD_GO"}, "postgres"),
		getEnvAny([]string{"DB_NAME", "DB_NAME_GO"}, "doctor_booking"),
		getEnvAny([]string{"DB_PORT", "DB_PORT_GO"}, "5432"),
		getEnvAny([]string{"DB_SSLMODE", "DB_SSLMODE_GO"}, "disable"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.Hospital{},
		&models.Specialization{},
		&models.Doctor{},
		&models.DoctorSchedule{},
		&models.User{},
		&models.Appointment{},
	)
	if err != nil {
		return nil, fmt.Errorf("auto-migration failed: %w", err)
	}

	return db, nil
}

func getEnvAny(keys []string, fallback string) string {
	for _, key := range keys {
		if val := os.Getenv(key); val != "" {
			return val
		}
	}
	return fallback
}
