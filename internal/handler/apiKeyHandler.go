package handler

import (
	"fmt"
	"strings"
	"time"

	"github.com/SwanHtetAungPhyo/kyc-api/pkg/config"
	"github.com/SwanHtetAungPhyo/kyc-api/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type APIKeyHandler struct {
	logger    logger.Logger
	jwtSecret string
}

func NewAPIKeyHandler(log logger.Logger, cfg *config.Config) *APIKeyHandler {
	return &APIKeyHandler{
		logger:    log,
		jwtSecret: cfg.JWT.Secret,
	}
}

func (h *APIKeyHandler) GenerateAPIKey(c *fiber.Ctx) error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(30 * 24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate API key")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to generate API key",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"api_key": tokenString,
		"expires": time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339),
	})
}

func (h *APIKeyHandler) JWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Missing or invalid Authorization header",
			})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(h.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			h.logger.WithError(err).Error("Invalid or expired API key")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid or expired API key",
			})
		}

		return c.Next()
	}
}

func (h *APIKeyHandler) RegisterRoutes(app *fiber.App) {
	app.Post("/api-key", h.GenerateAPIKey)
}
