package main

import (
	"context"
	"log"
	"os" // Added to access environment variables directly if needed

	"service-antrik-chatbot/config"
	"service-antrik-chatbot/controllers"
	"service-antrik-chatbot/repository"
	"service-antrik-chatbot/routes"

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
    // In main.go
    err := godotenv.Load("../.env") // It's now one level up from the binary
    if err != nil {
        log.Println("Note: .env file not found, using system environment variables")
    }

    // 2. Initialize DB
    db, err := config.ConnectDatabase()
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }

    redisClient := redis.NewClient(&redis.Options{
        Addr:     "192.168.110.102:6379", // Change if your n8n Redis is on a different host/port
        Password: "",               // Add your Redis password if you have one
        DB:       0,                // Default DB is usually 0
    })

    // Optional: Ping Redis to ensure the connection works before starting the app
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

    // Setup Router
    r := gin.Default()

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

    // 3. Setup Port
    port := getEnv("APP_PORT", "8080")
    addr := ":" + port

    log.Printf("Server running on %s", addr)
    
    // 4. Start Server
    if err := r.Run(addr); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}