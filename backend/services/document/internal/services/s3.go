package services

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/keshvan/rag-as-service/backend/services/document/internal/config"
)

type S3PresignService struct {
	presigner *s3.PresignClient
	bucket    string
	ttl       time.Duration
}

func NewS3PresignService(ctx context.Context, cfg *config.DocumentConfig) (*S3PresignService, error) {
	awsCfg, err := awscfg.LoadDefaultConfig(
		ctx,
		awscfg.WithRegion(cfg.S3Region),
		awscfg.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.S3AccessKeyID, cfg.S3SecretAccessKey, ""),
		),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.S3Endpoint)
		o.UsePathStyle = true
	})

	return &S3PresignService{
		presigner: s3.NewPresignClient(client),
		bucket:    cfg.S3Bucket,
		ttl:       time.Duration(cfg.S3PresignTTLSecond) * time.Second,
	}, nil
}

func (s *S3PresignService) PresignPut(ctx context.Context, objectKey, contentType string) (string, map[string]string, int32, error) {
	out, err := s.presigner.PresignPutObject(
		ctx,
		&s3.PutObjectInput{
			Bucket:      aws.String(s.bucket),
			Key:         aws.String(objectKey),
			ContentType: aws.String(contentType),
		},
		s3.WithPresignExpires(s.ttl),
	)
	if err != nil {
		return "", nil, 0, err
	}

	return out.URL, map[string]string{"Content-Type": contentType}, int32(s.ttl.Seconds()), nil
}

func (s *S3PresignService) BuildObjectKey(orgID, documentID, fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	if ext == "" {
		ext = ".bin"
	}
	return fmt.Sprintf("%s/documents/%s/source%s", orgID, documentID, ext)
}
