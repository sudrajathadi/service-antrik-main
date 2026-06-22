package main

import (
	"context"
	"log"
	"os"
	"time" // Added for CORS MaxAge

	"service-antrik-chatbot/config"
	"service-antrik-chatbot/controllers"
	"service-antrik-chatbot/repository"
	"service-antrik-chatbot/routes"

	"github.com/gin-contrib/cors" // Added CORS package
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

// Helper function if not already available globally
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	// 1. Load .env FIRST
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("Note: .env file not found, using system environment variables")
	}

	// 2. Initialize DB
	db, err := config.ConnectDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	redisHost := getEnv("REDIS_HOST", "192.168.110.102")
	redisPort := getEnv("REDIS_PORT", "6379")
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       0,
	})

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Initialize Repositories
	hospitalRepo := repository.NewHospitalRepository(db)
	specializationRepo := repository.NewSpecializationRepository(db)
	doctorRepo := repository.NewDoctorRepository(db)
	scheduleRepo := repository.NewDoctorScheduleRepository(db)
	userRepo := repository.NewUserRepository(db, redisClient)
	appointmentRepo := repository.NewAppointmentRepository(db)

	// Initialize Controllers
	hospitalCtrl := controllers.NewHospitalController(hospitalRepo)
	specializationCtrl := controllers.NewSpecializationController(specializationRepo)
	doctorCtrl := controllers.NewDoctorController(doctorRepo)
	scheduleCtrl := controllers.NewDoctorScheduleController(scheduleRepo)
	userCtrl := controllers.NewUserController(userRepo)
	appointmentCtrl := controllers.NewAppointmentController(appointmentRepo)
	bulkUploadCtrl := controllers.NewBulkUploadController(db)

	// Setup Router
	r := gin.Default()

	// --- NEW CORS CONFIGURATION ---
	r.Use(cors.New(cors.Config{
		// Allow both HTTP and HTTPS for your domain, plus localhost for frontend development
		AllowOrigins: []string{
			"http://chatbot.sederajat.work",
			"https://chatbot.sederajat.work",
			"http://localhost:3000",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	// ------------------------------

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "AIR WORKING V1",
		})
	})

	routes.RegisterHospitalRoutes(r, hospitalCtrl)
	routes.RegisterSpecializationRoutes(r, specializationCtrl)
	routes.RegisterDoctorRoutes(r, doctorCtrl)
	routes.RegisterDoctorScheduleRoutes(r, scheduleCtrl)
	routes.RegisterUserRoutes(r, userCtrl)
	routes.RegisterAppointmentRoutes(r, appointmentCtrl)
	routes.RegisterBulkUploadRoutes(r, bulkUploadCtrl)

	// 3. Setup Port
	port := getEnv("APP_PORT", "8080")
	addr := ":" + port

	log.Printf("Server running on %s", addr)

	// 4. Start Server
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
