package main

import (
	"log"

	"github.com/SwanHtetAungPhyo/kyc-api/internal/handler"
	"github.com/SwanHtetAungPhyo/kyc-api/internal/repo"
	"github.com/SwanHtetAungPhyo/kyc-api/internal/service"
	"github.com/SwanHtetAungPhyo/kyc-api/pkg/config"
	"github.com/SwanHtetAungPhyo/kyc-api/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log := logger.NewLogger()
	log.Info("Starting KYC verification service")

	// Initialize AWS repository
	awsRepo, err := repo.NewAWSRepository(
		cfg.AWS.AccessKeyID,
		cfg.AWS.SecretAccessKey,
		cfg.AWS.Region,
	)
	if err != nil {
		log.WithError(err).Error("Failed to initialize AWS repository")
		return
	}

	// Initialize services
	kycService := service.NewKYCService(awsRepo, log)

	// Initialize handlers
	kycHandler := handler.NewKYCHandler(kycService, log)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			log.WithError(err).Error("Request failed")

			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		},
	})

	// Configure CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "*",
		AllowHeaders: "*",
	}))

	// Register routes
	kycHandler.RegisterRoutes(app)

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "kyc-verification",
		})
	})

	// Start server
	port := ":" + cfg.Server.Port
	log.WithField("port", cfg.Server.Port).Info("Server starting")

	if err := app.Listen(port); err != nil {
		log.WithError(err).Error("Failed to start server")
	}
}
