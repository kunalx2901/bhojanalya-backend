package storage

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2Client struct {
	client  *s3.Client
	bucket  string
	baseURL string
}

func NewR2Client(ctx context.Context) (*R2Client, error) {
	endpoint := os.Getenv("R2_ENDPOINT")
	accessKey := os.Getenv("R2_ACCESS_KEY")
	secretKey := os.Getenv("R2_SECRET_KEY")
	bucket := os.Getenv("R2_BUCKET_NAME")
	baseURL := os.Getenv("R2_PUBLIC_BASE_URL")

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion("auto"),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				accessKey,
				secretKey,
				"",
			),
		),
		config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...interface{}) (aws.Endpoint, error) {
					if service == s3.ServiceID {
						return aws.Endpoint{
							URL:           endpoint,
							SigningRegion: "auto",
						}, nil
					}
					return aws.Endpoint{}, &aws.EndpointNotFoundError{}
				},
			),
		),
	)
	if err != nil {
		return nil, err
	}

	return &R2Client{
		client: s3.NewFromConfig(cfg),
		bucket: bucket,
		baseURL: baseURL,
	}, nil
}

func (r *R2Client) Upload(ctx context.Context, key string, file multipart.File) (string, error) {
	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &r.bucket,
		Key:    &key,
		Body:   file,
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", r.baseURL, key), nil
}
