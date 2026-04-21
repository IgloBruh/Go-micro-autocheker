package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Store struct {
	Client *minio.Client
}

func New(endpoint, accessKey, secretKey string, useSSL bool) (*Store, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	return &Store{Client: client}, nil
}

func (s *Store) EnsureBucket(ctx context.Context, bucket string) error {
	exists, err := s.Client.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return s.Client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
}

func (s *Store) PutObject(ctx context.Context, bucket, key string, data []byte, contentType string) error {
	reader := bytes.NewReader(data)
	_, err := s.Client.PutObject(ctx, bucket, key, reader, int64(len(data)), minio.PutObjectOptions{ContentType: contentType})
	return err
}

func (s *Store) PutJSON(ctx context.Context, bucket, key string, v any) error {
	payload, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return s.PutObject(ctx, bucket, key, payload, "application/json")
}

func (s *Store) GetObject(ctx context.Context, bucket, key string) ([]byte, error) {
	obj, err := s.Client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()
	return io.ReadAll(obj)
}

func (s *Store) ListKeys(ctx context.Context, bucket, prefix string) ([]string, error) {
	var result []string
	for object := range s.Client.ListObjects(ctx, bucket, minio.ListObjectsOptions{Prefix: prefix, Recursive: true}) {
		if object.Err != nil {
			return nil, object.Err
		}
		if filepath.Base(object.Key) == ".keep" {
			continue
		}
		result = append(result, object.Key)
	}
	return result, nil
}

func (s *Store) MustGetJSON(ctx context.Context, bucket, key string, dst any) error {
	payload, err := s.GetObject(ctx, bucket, key)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(payload, dst); err != nil {
		return fmt.Errorf("decode %s/%s: %w", bucket, key, err)
	}
	return nil
}
