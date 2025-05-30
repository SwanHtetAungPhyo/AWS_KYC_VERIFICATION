package repo

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	rtype "github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	"github.com/aws/aws-sdk-go-v2/service/textract"
	textraTyp "github.com/aws/aws-sdk-go-v2/service/textract/types"
)

// AWSRepository handles AWS service interactions
type AWSRepository interface {
	AnalyzeID(ctx context.Context, idBlob []byte) (*textract.AnalyzeIDOutput, error)
	DetectFaces(ctx context.Context, imageBlob []byte) (*rekognition.DetectFacesOutput, error)
	CompareFaces(ctx context.Context, srcBlob, targetBlob []byte, threshold float32) (*rekognition.CompareFacesOutput, error)
}

// awsRepository implements AWSRepository
type awsRepository struct {
	textractClient    *textract.Client
	rekognitionClient *rekognition.Client
}

// NewAWSRepository creates a new AWS repository instance
func NewAWSRepository(accessKey, secretKey, region string) (AWSRepository, error) {
	ctx := context.Background()

	customCfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &awsRepository{
		textractClient:    textract.NewFromConfig(customCfg),
		rekognitionClient: rekognition.NewFromConfig(customCfg),
	}, nil
}

// AnalyzeID analyzes an ID document using AWS Textract
func (r *awsRepository) AnalyzeID(ctx context.Context, idBlob []byte) (*textract.AnalyzeIDOutput, error) {
	input := &textract.AnalyzeIDInput{
		DocumentPages: []textraTyp.Document{
			{Bytes: idBlob},
		},
	}

	result, err := r.textractClient.AnalyzeID(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("textract analysis failed: %w", err)
	}

	return result, nil
}

// DetectFaces detects faces in an image using AWS Rekognition
func (r *awsRepository) DetectFaces(ctx context.Context, imageBlob []byte) (*rekognition.DetectFacesOutput, error) {
	input := &rekognition.DetectFacesInput{
		Image: &rtype.Image{
			Bytes: imageBlob,
		},
		Attributes: []rtype.Attribute{
			rtype.AttributeDefault,
			rtype.AttributeAll,
		},
	}

	result, err := r.rekognitionClient.DetectFaces(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("face detection failed: %w", err)
	}

	return result, nil
}

// CompareFaces compares two faces using AWS Rekognition
func (r *awsRepository) CompareFaces(ctx context.Context, srcBlob, targetBlob []byte, threshold float32) (*rekognition.CompareFacesOutput, error) {
	input := &rekognition.CompareFacesInput{
		SourceImage: &rtype.Image{
			Bytes: srcBlob,
		},
		TargetImage: &rtype.Image{
			Bytes: targetBlob,
		},
		SimilarityThreshold: aws.Float32(threshold),
	}

	result, err := r.rekognitionClient.CompareFaces(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("face comparison failed: %w", err)
	}

	return result, nil
}
