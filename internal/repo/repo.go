package repo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/SwanHtetAungPhyo/kyc-api/internal/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	rtype "github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	"github.com/aws/aws-sdk-go-v2/service/textract"
	textraTyp "github.com/aws/aws-sdk-go-v2/service/textract/types"
)

type AWSRepository interface {
	AnalyzeID(ctx context.Context, idBlob []byte) (*textract.AnalyzeIDOutput, error)
	DetectFaces(ctx context.Context, imageBlob []byte) (*rekognition.DetectFacesOutput, error)
	CompareFaces(ctx context.Context, srcBlob, targetBlob []byte, threshold float32) (*rekognition.CompareFacesOutput, error)
	RecordAttempt(ctx context.Context, email string, success bool) error
	CheckIfProceed(ctx context.Context, email string) (bool, error)
}

type awsRepository struct {
	textractClient    *textract.Client
	rekognitionClient *rekognition.Client
	dynamoDBClient    *dynamodb.Client
}

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
		dynamoDBClient:    dynamodb.NewFromConfig(customCfg),
	}, nil
}

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

func (r *awsRepository) RecordAttempt(ctx context.Context, email string, success bool) error {
	record := models.EmailRecord{
		Email:       email,
		AttemptedAt: time.Now(),
		Processed:   success,
	}

	item, err := attributevalue.MarshalMap(record)
	if err != nil {
		return fmt.Errorf("failed to put the item: %s", err.Error())
	}
	tableName := os.Getenv("KYC_RECORD")

	_, err = r.dynamoDBClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           &tableName,
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(email)"),
	})

	var ccfe *types.ConditionalCheckFailedException
	if errors.Is(err, ccfe) {
		return errors.New("already processed")
	}
	return nil
}

func (r *awsRepository) CheckIfProceed(ctx context.Context, email string) (bool, error) {
	tableName := os.Getenv("KYC_RECORD")
	result, err := r.dynamoDBClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"email": &types.AttributeValueMemberS{
				Value: email,
			},
		},
	})
	if err != nil {
		return false, fmt.Errorf("failed to get item: %s", err.Error())
	}

	if result.Item == nil {
		return false, nil
	}

	var record models.EmailRecord
	if err := attributevalue.UnmarshalMap(result.Item, &record); err != nil {
		return false, fmt.Errorf("failed to unmarshal data: %s", err.Error())
	}
	return record.Processed, nil
}
