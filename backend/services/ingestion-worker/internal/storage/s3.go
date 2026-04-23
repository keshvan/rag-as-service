package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Downloader struct {
	client      *s3.Client
	bucket      string
	downloadDir string
}

func NewS3Downloader(
	ctx context.Context,
	endpoint, region, bucket, accessKeyID, secretAccessKey, downloadDir string,
) (*S3Downloader, error) {
	awsCfg, err := awscfg.LoadDefaultConfig(
		ctx,
		awscfg.WithRegion(region),
		awscfg.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	return &S3Downloader{
		client:      client,
		bucket:      bucket,
		downloadDir: downloadDir,
	}, nil
}

func (d *S3Downloader) DownloadToFile(ctx context.Context, objectKey, orgID, docID string) (string, error) {
	out, err := d.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return "", fmt.Errorf("get object from s3: %w", err)
	}
	defer out.Body.Close()

	targetDir := filepath.Join(d.downloadDir, orgID, docID)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return "", fmt.Errorf("create target dir: %w", err)
	}

	fileName := filepath.Base(objectKey)
	targetPath := filepath.Join(targetDir, fileName)

	file, err := os.Create(targetPath)
	if err != nil {
		return "", fmt.Errorf("create local file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, out.Body); err != nil {
		return "", fmt.Errorf("copy object body to file: %w", err)
	}

	return targetPath, nil
}
