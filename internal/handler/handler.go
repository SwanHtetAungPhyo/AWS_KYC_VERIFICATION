package handler

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/SwanHtetAungPhyo/kyc-api/internal/models"
	"github.com/SwanHtetAungPhyo/kyc-api/internal/service"
	"github.com/SwanHtetAungPhyo/kyc-api/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

type KYCHandler struct {
	kycService service.KYCService
	logger     logger.Logger
}

func NewKYCHandler(kycService service.KYCService, log logger.Logger) *KYCHandler {
	return &KYCHandler{
		kycService: kycService,
		logger:     log,
	}
}

func (h *KYCHandler) HandleKYCVerification(c *fiber.Ctx) error {
	var req models.KYCRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WithError(err).Error("Failed to parse request body")
		return c.Status(fiber.StatusBadRequest).JSON(models.KYCResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse request body: %v", err),
		})
	}

	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.KYCResponse{
			Success: false,
			Error:   "Email is required",
		})
	}

	proceed, err := h.kycService.CheckIfProceed(c.Context(), req.Email)
	if err != nil {
		h.logger.WithError(err).Error("Failed to check email status")
		return c.Status(fiber.StatusInternalServerError).JSON(models.KYCResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to check email status: %v", err),
		})
	}
	if proceed {
		return c.Status(fiber.StatusBadRequest).JSON(models.KYCResponse{
			Success: false,
			Error:   "KYC with this email is already done successfully",
		})
	}

	idBlob, err := h.getFileBlob(c, "id_image")
	if err != nil {
		h.logger.WithError(err).Error("Failed to process ID image")
		return c.Status(fiber.StatusBadRequest).JSON(models.KYCResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to process ID image: %v", err),
		})
	}

	selfieBlob, err := h.getFileBlob(c, "selfie")
	if err != nil {
		h.logger.WithError(err).Error("Failed to process selfie")
		return c.Status(fiber.StatusBadRequest).JSON(models.KYCResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to process selfie: %v", err),
		})
	}

	result, err := h.kycService.VerifyKYC(c.Context(), idBlob, selfieBlob, req.Email)
	if err != nil {
		h.logger.WithError(err).Error("KYC verification failed")
		return c.Status(fiber.StatusBadRequest).JSON(models.KYCResponse{
			Success: false,
			Error:   fmt.Sprintf("KYC verification failed: %v", err),
		})
	}

	response := models.KYCResponse{
		Success:    true,
		Verified:   result.Verified,
		Similarity: result.Similarity,
		Message:    result.Message,
	}

	h.logger.Info("KYC response sent", "response", response)
	return c.JSON(response)
}

func (h *KYCHandler) getFileBlob(c *fiber.Ctx, fieldName string) ([]byte, error) {
	fileHeader, err := c.FormFile(fieldName)
	if err != nil {
		return nil, fmt.Errorf("missing or invalid file: %w", err)
	}

	if !strings.HasPrefix(fileHeader.Header.Get("Content-Type"), "image/") {
		return nil, errors.New("file must be an image (JPG, PNG)")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	blob, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return blob, nil
}

func (h *KYCHandler) RegisterRoutes(app *fiber.App) {
	app.Post("/kyc", h.HandleKYCVerification)
}
