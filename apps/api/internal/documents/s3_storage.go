package documents

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3StorageConfig struct {
	Bucket       string
	Region       string
	BaseEndpoint string
	UsePathStyle bool
}

type S3Storage struct {
	bucket    string
	client    *s3.Client
	presigner *s3.PresignClient
}

func NewS3Storage(ctx context.Context, cfg S3StorageConfig) (*S3Storage, error) {
	cfg.Bucket = strings.TrimSpace(cfg.Bucket)
	cfg.Region = strings.TrimSpace(cfg.Region)
	if cfg.Bucket == "" || cfg.Region == "" {
		return nil, errors.New("storage bucket and region are required")
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("load AWS configuration: %w", err)
	}
	client := s3.NewFromConfig(awsCfg, func(options *s3.Options) {
		options.UsePathStyle = cfg.UsePathStyle
		if strings.TrimSpace(cfg.BaseEndpoint) != "" {
			options.BaseEndpoint = aws.String(strings.TrimSpace(cfg.BaseEndpoint))
		}
	})
	return &S3Storage{bucket: cfg.Bucket, client: client, presigner: s3.NewPresignClient(client)}, nil
}

func (s *S3Storage) PresignPut(ctx context.Context, key, contentType string, expires time.Duration) (string, map[string]string, error) {
	result, err := s.presigner.PresignPutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key), ContentType: aws.String(contentType)}, func(options *s3.PresignOptions) { options.Expires = expires })
	if err != nil {
		return "", nil, err
	}
	headers := make(map[string]string)
	for key, values := range result.SignedHeader {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	return result.URL, headers, nil
}

func (s *S3Storage) Head(ctx context.Context, key string) (ObjectInfo, error) {
	result, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	if err != nil {
		return ObjectInfo{}, err
	}
	return ObjectInfo{SizeBytes: aws.ToInt64(result.ContentLength), ContentType: aws.ToString(result.ContentType), ETag: aws.ToString(result.ETag)}, nil
}

func (s *S3Storage) PresignGet(ctx context.Context, key string, expires time.Duration) (string, error) {
	result, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)}, func(options *s3.PresignOptions) { options.Expires = expires })
	if err != nil {
		return "", err
	}
	return result.URL, nil
}

func (s *S3Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	return err
}
