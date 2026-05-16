package cloud

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/cenkalti/backoff/v4"
)

// R2BackupService handles backup operations to Cloudflare R2
type R2BackupService struct {
	client     *s3.Client
	bucket     string
	accountID  string
	region     string
	maxRetries int
}

// R2Config holds R2 configuration
type R2Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	Region          string
}

// NewR2BackupService creates a new R2 backup service
func NewR2BackupService(ctx context.Context, cfg R2Config) (*R2BackupService, error) {
	if cfg.AccountID == "" || cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" || cfg.Bucket == "" {
		return nil, fmt.Errorf("R2 configuration is incomplete")
	}

	// Set default region if not specified
	region := cfg.Region
	if region == "" || region == "auto" {
		region = "auto"
	}

	// Create AWS configuration for R2
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     cfg.AccessKeyID,
				SecretAccessKey: cfg.SecretAccessKey,
			}, nil
		})),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client with R2 endpoint
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		// Cloudflare R2 endpoint
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID))
	})

	return &R2BackupService{
		client:     client,
		bucket:     cfg.Bucket,
		accountID:  cfg.AccountID,
		region:     region,
		maxRetries: 3,
	}, nil
}

// UploadBackup uploads a local backup file to R2 with retry logic
func (s *R2BackupService) UploadBackup(ctx context.Context, localPath, remoteKey string) error {
	// Open the local file
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	// Get file info for content length
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Create upload with retry logic
	var lastErr error
	operation := func() error {
		// Reset file pointer to beginning
		_, err := file.Seek(0, 0)
		if err != nil {
			return fmt.Errorf("failed to seek file: %w", err)
		}

		// Upload to R2
		size := fileInfo.Size()
		_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:        aws.String(s.bucket),
			Key:           aws.String(remoteKey),
			Body:          file,
			ContentLength: &size,
		})
		if err != nil {
			return fmt.Errorf("failed to upload to R2: %w", err)
		}

		return nil
	}

	// Configure exponential backoff
	backoffCfg := backoff.NewExponentialBackOff()
	backoffCfg.InitialInterval = 1 * time.Second
	backoffCfg.MaxInterval = 30 * time.Second
	backoffCfg.MaxElapsedTime = 2 * time.Minute
	backoffCfg.Multiplier = 2.0

	// Execute with retry
	err = backoff.Retry(operation, backoff.WithMaxRetries(backoffCfg, uint64(s.maxRetries)))
	if err != nil {
		lastErr = err
		return fmt.Errorf("upload failed after %d retries: %w", s.maxRetries, lastErr)
	}

	return nil
}

// DownloadBackup downloads a backup file from R2
func (s *R2BackupService) DownloadBackup(ctx context.Context, remoteKey, localPath string) error {
	// Get object from R2
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(remoteKey),
	})
	if err != nil {
		return fmt.Errorf("failed to download from R2: %w", err)
	}
	defer result.Body.Close()

	// Create local file
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	// Copy content
	_, err = io.Copy(file, result.Body)
	if err != nil {
		return fmt.Errorf("failed to write file content: %w", err)
	}

	return nil
}

// ListBackups lists all backups in the R2 bucket
func (s *R2BackupService) ListBackups(ctx context.Context, prefix string) ([]string, error) {
	result, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	var keys []string
	for _, obj := range result.Contents {
		keys = append(keys, *obj.Key)
	}

	return keys, nil
}

// DeleteBackup deletes a backup from R2
func (s *R2BackupService) DeleteBackup(ctx context.Context, remoteKey string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(remoteKey),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}
