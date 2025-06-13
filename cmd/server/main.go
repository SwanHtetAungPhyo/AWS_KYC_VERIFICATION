package main

import (
	"log"
	"time"

	"github.com/SwanHtetAungPhyo/kyc-api/internal/handler"
	"github.com/SwanHtetAungPhyo/kyc-api/internal/repo"
	"github.com/SwanHtetAungPhyo/kyc-api/internal/service"
	"github.com/SwanHtetAungPhyo/kyc-api/pkg/config"
	"github.com/SwanHtetAungPhyo/kyc-api/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log := logger.NewLogger()
	log.Info("Starting KYC verification service")

	awsRepo, err := repo.NewAWSRepository(
		cfg.AWS.AccessKeyID,
		cfg.AWS.SecretAccessKey,
		cfg.AWS.Region,
	)
	if err != nil {
		log.WithError(err).Error("Failed to initialize AWS repository")
		return
	}

	kycService := service.NewKYCService(awsRepo, log)
	kycHandler := handler.NewKYCHandler(kycService, log)

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

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "*",
		AllowHeaders: "*",
	}))

	app.Use(limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"error":   "Rate limit exceeded. Please try again later.",
			})
		},
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "kyc-verification",
		})
	})

	kycHandler.RegisterRoutes(app)

	port := ":" + cfg.Server.Port
	log.WithField("port", cfg.Server.Port).Info("Server starting")

	if err := app.Listen(port); err != nil {
		log.WithError(err).Error("Failed to start server")
	}
}
