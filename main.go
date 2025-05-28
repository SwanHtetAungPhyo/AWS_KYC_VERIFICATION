package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	rtype "github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	"github.com/aws/aws-sdk-go-v2/service/textract"
	textraTyp "github.com/aws/aws-sdk-go-v2/service/textract/types"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type KYCRequest struct {
	Email string `form:"email"`
}

type KYCResponse struct {
	Success    bool    `json:"success"`
	Verified   bool    `json:"verified"`
	Similarity float32 `json:"similarity,omitempty"`
	Message    string  `json:"message"`
	Error      string  `json:"error,omitempty"`
}

type Service struct {
	textractClient    *textract.Client
	rekognitionClient *rekognition.Client
	log               *logrus.Logger
	ctx               context.Context
}

func NewService(ctx context.Context) (*Service, error) {
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	region := os.Getenv("AWS_REGION")

	if accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("AWS credentials not set in environment variables")
	}

	customCfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	return &Service{
		textractClient:    textract.NewFromConfig(customCfg),
		rekognitionClient: rekognition.NewFromConfig(customCfg),
		log:               logger,
		ctx:               ctx,
	}, nil
}

func (s *Service) HandleKYCVerification(c *fiber.Ctx) error {
	var req KYCRequest
	if err := c.BodyParser(&req); err != nil {
		s.log.WithError(err).Error("Failed to parse request body")
		return c.Status(fiber.StatusBadRequest).JSON(KYCResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(KYCResponse{
			Success: false,
			Error:   "Email is required",
		})
	}

	idImage, err := c.FormFile("id_image")
	if err != nil {
		s.log.WithError(err).Error("Failed to get ID image")
		return c.Status(fiber.StatusBadRequest).JSON(KYCResponse{
			Success: false,
			Error:   "ID image is required",
		})
	}

	selfie, err := c.FormFile("selfie")
	if err != nil {
		s.log.WithError(err).Error("Failed to get selfie")
		return c.Status(fiber.StatusBadRequest).JSON(KYCResponse{
			Success: false,
			Error:   "Selfie is required",
		})
	}

	idFile, err := idImage.Open()
	if err != nil {
		s.log.WithError(err).Error("Failed to open ID image")
		return c.Status(fiber.StatusInternalServerError).JSON(KYCResponse{
			Success: false,
			Error:   "Failed to process ID image",
		})
	}
	defer idFile.Close()

	selfieFile, err := selfie.Open()
	if err != nil {
		s.log.WithError(err).Error("Failed to open selfie")
		return c.Status(fiber.StatusInternalServerError).JSON(KYCResponse{
			Success: false,
			Error:   "Failed to process selfie",
		})
	}
	defer selfieFile.Close()

	idBlob, err := io.ReadAll(idFile)
	if err != nil {
		s.log.WithError(err).Error("Failed to read ID image")
		return c.Status(fiber.StatusInternalServerError).JSON(KYCResponse{
			Success: false,
			Error:   "Failed to read ID image",
		})
	}

	selfieBlob, err := io.ReadAll(selfieFile)
	if err != nil {
		s.log.WithError(err).Error("Failed to read selfie")
		return c.Status(fiber.StatusInternalServerError).JSON(KYCResponse{
			Success: false,
			Error:   "Failed to read selfie",
		})
	}

	verified, similarity, err := s.KYCVerification(idBlob, selfieBlob, req.Email)
	if err != nil {
		s.log.WithError(err).Error("KYC verification failed")
		return c.Status(fiber.StatusInternalServerError).JSON(KYCResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	response := KYCResponse{
		Success:    true,
		Verified:   verified,
		Similarity: similarity,
		Message:    "KYC verification completed",
	}

	return c.JSON(response)
}

func (s *Service) KYCVerification(idBlob, selfieBlob []byte, email string) (bool, float32, error) {
	s.log.WithField("email", email).Info("Processing KYC verification")

	if len(idBlob) == 0 || len(selfieBlob) == 0 {
		s.log.Error("Empty blob data provided")
		return false, 0, errors.New("empty image data provided")
	}

	_, err := s.TextractClientProcessor(idBlob)
	if err != nil {
		s.log.WithError(err).Error("ID analysis failed")
		return false, 0, fmt.Errorf("ID analysis failed: %w", err)
	}

	faces, err := s.DetectFaces(selfieBlob)
	if err != nil {
		s.log.WithError(err).Error("Face detection failed")
		return false, 0, fmt.Errorf("face detection failed: %w", err)
	}

	if err := s.Processing(faces, &textract.AnalyzeIDInput{
		DocumentPages: []textraTyp.Document{{Bytes: idBlob}},
	}); err != nil {
		s.log.WithError(err).Error("Face validation failed")
		return false, 0, fmt.Errorf("face validation failed: %w", err)
	}

	compareResult, err := s.CompareFaces(idBlob, selfieBlob)
	if err != nil {
		s.log.WithError(err).Error("Face comparison failed")
		return false, 0, fmt.Errorf("face comparison failed: %w", err)
	}

	if len(compareResult.FaceMatches) == 0 {
		s.log.Error("No face matches found")
		return false, 0, errors.New("no face matches found")
	}

	similarity := compareResult.FaceMatches[0].Similarity
	s.log.WithField("similarity", *similarity).Info("Face comparison result")

	verified := *similarity >= 70

	s.log.WithFields(logrus.Fields{
		"email":      email,
		"verified":   verified,
		"similarity": *similarity,
	}).Info("KYC verification completed")

	return verified, *similarity, nil
}

func (s *Service) Processing(faces *rekognition.DetectFacesOutput, src *textract.AnalyzeIDInput) error {
	if len(faces.FaceDetails) != 1 {
		s.log.WithField("face_count", len(faces.FaceDetails)).Error("Invalid number of faces detected")
		return errors.New("exactly one face should be detected")
	}

	face := faces.FaceDetails[0]
	minConfidence := float32(90)

	if face.Confidence == nil || *face.Confidence < minConfidence {
		s.log.WithField("confidence", *face.Confidence).Error("Low face detection confidence")
		return fmt.Errorf("low face detection confidence: %.2f", *face.Confidence)
	}

	if face.Quality == nil {
		s.log.Error("Missing face quality data")
		return errors.New("missing face quality data")
	}

	if face.Quality.Brightness == nil || face.Quality.Sharpness == nil {
		s.log.Error("Incomplete face quality metrics")
		return errors.New("incomplete face quality metrics")
	}

	minBrightness := float32(50)
	minSharpness := float32(50)
	if *face.Quality.Brightness < minBrightness || *face.Quality.Sharpness < minSharpness {
		s.log.WithFields(logrus.Fields{
			"brightness": *face.Quality.Brightness,
			"sharpness":  *face.Quality.Sharpness,
		}).Error("Poor image quality")
		return fmt.Errorf("poor image quality (brightness: %.2f, sharpness: %.2f)",
			*face.Quality.Brightness, *face.Quality.Sharpness)
	}

	s.log.WithFields(logrus.Fields{
		"confidence": *face.Confidence,
		"brightness": *face.Quality.Brightness,
		"sharpness":  *face.Quality.Sharpness,
	}).Info("Face validation passed")
	return nil
}

func (s *Service) TextractClientProcessor(idBlob []byte) (*textract.AnalyzeIDOutput, error) {
	input := &textract.AnalyzeIDInput{
		DocumentPages: []textraTyp.Document{
			{Bytes: idBlob},
		},
	}

	result, err := s.textractClient.AnalyzeID(s.ctx, input)
	if err != nil {
		s.log.WithError(err).Error("Textract analysis failed")
		return nil, fmt.Errorf("textract analysis failed: %w", err)
	}

	s.log.WithField("result", result).Debug("Textract analysis completed")
	return result, nil
}

func (s *Service) DetectFaces(imageBlob []byte) (*rekognition.DetectFacesOutput, error) {
	input := &rekognition.DetectFacesInput{
		Image: &rtype.Image{
			Bytes: imageBlob,
		},
		Attributes: []rtype.Attribute{
			rtype.AttributeDefault,
			rtype.AttributeAll,
		},
	}

	result, err := s.rekognitionClient.DetectFaces(s.ctx, input)
	if err != nil {
		s.log.WithError(err).Error("Face detection failed")
		return nil, fmt.Errorf("face detection failed: %w", err)
	}

	s.log.WithField("face_count", len(result.FaceDetails)).Debug("Face detection completed")
	return result, nil
}

func (s *Service) CompareFaces(srcBlob, targetBlob []byte) (*rekognition.CompareFacesOutput, error) {
	input := &rekognition.CompareFacesInput{
		SourceImage: &rtype.Image{
			Bytes: srcBlob,
		},
		TargetImage: &rtype.Image{
			Bytes: targetBlob,
		},
		SimilarityThreshold: aws.Float32(70.0),
	}

	result, err := s.rekognitionClient.CompareFaces(s.ctx, input)
	if err != nil {
		s.log.WithError(err).Error("Face comparison failed")
		return nil, fmt.Errorf("face comparison failed: %w", err)
	}

	s.log.WithField("matches", len(result.FaceMatches)).Debug("Face comparison completed")
	return result, nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println(err.Error())
		return
	}
	service, err := NewService(context.Background())
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize service: %v", err))
	}

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "*",
		AllowHeaders: "*",
	}))
	app.Post("/kyc", service.HandleKYCVerification)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	app.Listen(":" + port)
}
