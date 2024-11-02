package main

import (
	"badbuddy/internal/delivery/http/rest"
	"badbuddy/internal/infrastructure/database"
	"badbuddy/internal/infrastructure/server"
	"badbuddy/internal/repositories/postgres"
	"badbuddy/internal/usecase/session"
	"badbuddy/internal/usecase/user"
	"badbuddy/internal/usecase/venue"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Println("Warning: No .env file found")
	}

	// Now that env vars are loaded, we can use getEnv
	fmt.Println("badbuddy API", getEnv("DB_HOST", "beer"))

	// Create a new configuration
	dbConfig := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvAsInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", ""),
		DBName:   getEnv("DB_NAME", "general"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	db, err := database.NewSQLxDB(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.CloseSQLxDB(db)

	app := server.NewFiberServer()

	userRepo := postgres.NewUserRepository(db)
	userUseCase := user.NewUserUseCase(userRepo, "your-jwt-secret", 24*time.Hour)
	userHandler := rest.NewUserHandler(userUseCase)
	userHandler.SetupUserRoutes(app)

	venueRepo := postgres.NewVenueRepository(db)
	venueUseCase := venue.NewVenueUseCase(venueRepo, userRepo)
	venueHandler := rest.NewVenueHandler(venueUseCase)
	venueHandler.SetupVenueRoutes(app)

	sessionRepo := postgres.NewSessionRepository(db)
	sessionUseCase := session.NewSessionUseCase(sessionRepo, venueRepo)
	sessionHandler := rest.NewSessionHandler(sessionUseCase)
	sessionHandler.SetupSessionRoutes(app)

	//add heatlh check and ready check

	app.Get("*", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	port := getEnv("PORT", "8004")
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// Helper function to read an environment variable or return a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// Helper function to read an environment variable as an integer or return a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// Helper function to read an environment variable as a duration or return a default value
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}
