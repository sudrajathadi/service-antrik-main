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
		getEnv("DB_HOST_GO", "localhost"),
		getEnv("DB_USER_GO", "postgres"),
		getEnv("DB_PASSWORD_GO", "postgres"),
		getEnv("DB_NAME_GO", "doctor_booking"),
		getEnv("DB_PORT_GO", "5432"),
		getEnv("DB_SSLMODE_GO", "disable"),
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

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
