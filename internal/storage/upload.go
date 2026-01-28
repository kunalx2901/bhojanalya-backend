package storage

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// UploadMultipartFile uploads multipart file to R2 and returns public URL
func UploadMultipartFile(
	ctx context.Context,
	client *s3.Client,
	bucket string,
	key string,
	file *multipart.FileHeader,
) (string, error) {

	f, err := file.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()

	contentType := file.Header.Get("Content-Type")

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &bucket,
		Key:         &key,
		Body:        f,
		ContentType: &contentType,
	})
	if err != nil {
		return "", err
	}

	// Public R2 URL
	url := fmt.Sprintf("https://%s/%s", bucket, key)
	return url, nil
}
