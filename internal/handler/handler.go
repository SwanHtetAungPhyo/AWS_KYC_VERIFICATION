package handler

import (
	"io"

	"github.com/SwanHtetAungPhyo/kyc-api/internal/models"
	"github.com/SwanHtetAungPhyo/kyc-api/internal/repo"
	"github.com/SwanHtetAungPhyo/kyc-api/internal/service"
	"github.com/SwanHtetAungPhyo/kyc-api/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

type KYCHandler struct {
	kycService service.KYCService
	logger     logger.Logger
	repo       *repo.AWSRepository
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
			Error:   "Invalid request body",
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
			Error:   "Failed to check email status",
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
			Error:   "ID image is required and must be valid",
		})
	}

	selfieBlob, err := h.getFileBlob(c, "selfie")
	if err != nil {
		h.logger.WithError(err).Error("Failed to process selfie")
		return c.Status(fiber.StatusBadRequest).JSON(models.KYCResponse{
			Success: false,
			Error:   "Selfie is required and must be valid",
		})
	}

	result, err := h.kycService.VerifyKYC(c.Context(), idBlob, selfieBlob, req.Email)
	if err != nil {
		h.logger.WithError(err).Error("KYC verification failed")
		return c.Status(fiber.StatusInternalServerError).JSON(models.KYCResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	response := models.KYCResponse{
		Success:    true,
		Verified:   result.Verified,
		Similarity: result.Similarity,
		Message:    result.Message,
	}

	return c.JSON(response)
}

func (h *KYCHandler) getFileBlob(c *fiber.Ctx, fieldName string) ([]byte, error) {
	fileHeader, err := c.FormFile(fieldName)
	if err != nil {
		return nil, err
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	blob, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return blob, nil
}

func (h *KYCHandler) RegisterRoutes(app *fiber.App) {
	app.Post("/kyc", h.HandleKYCVerification)
}
