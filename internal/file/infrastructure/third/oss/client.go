package oss

import (
	"context"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/viper"
)

type Service interface {
	CheckBucketExists(ctx context.Context, bucketName string) (bool, error)
	CreateBucket(ctx context.Context, bucketName string) error
	CheckFileExists(ctx context.Context, bucketName, fileName string) (bool, error)
	PreUpload(ctx context.Context, file *Object) (*UploadResponse, error)
}

func NewService(conf *viper.Viper) Service {
	oss := conf.GetString("app.data.oss.driver")
	switch oss {
	case "minio":
		return NewMinioService(conf)
	// case "qiniu":
	// 	return NewQiniuService(conf)
	// case "tencent":
	// 	return NewTencentService(conf)
	// case "aliyun":
	// 	return NewAliyunService(conf)
	default:
		panic("unsupported oss driver")
	}
}

// func NewAliyunService(conf *viper.Viper) Service {
//
// }
//
// func NewTencentService(conf *viper.Viper) Service {
//
// }
//
// func NewQiniuService(conf *viper.Viper) Service {
//
// }

type minioService struct {
	minioClient *minio.Client
	expires     int64
}

func (m *minioService) CheckBucketExists(ctx context.Context, bucketName string) (bool, error) {
	exists, err := m.minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (m *minioService) CheckFileExists(ctx context.Context, bucketName, fileName string) (bool, error) {
	exists, err := m.minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}
	objectInfo, err := m.minioClient.StatObject(ctx, bucketName, fileName, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return objectInfo.ETag != "", nil
}

func (m *minioService) CreateBucket(ctx context.Context, bucketName string) error {
	err := m.minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		if exists, errBucketExists := m.minioClient.BucketExists(ctx, bucketName); errBucketExists == nil && exists {
			return nil
		}
		return err
	}
	return nil
}

func (m *minioService) PreUpload(ctx context.Context, file *Object) (*UploadResponse, error) {
	exists, err := m.CheckBucketExists(ctx, file.Bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		if err := m.CreateBucket(ctx, file.Bucket); err != nil {
			return nil, err
		}
	}
	presignedURL, err := m.minioClient.PresignedPutObject(ctx, file.Bucket, file.Key, time.Duration(m.expires))
	if err != nil {
		return nil, err
	}
	return &UploadResponse{
		UploadURL: presignedURL.String(),
		AccessURL: m.buildAccessURL(file.Bucket, file.Key),
		ExpiresAt: time.Now().Add(time.Duration(m.expires) * time.Second),
	}, nil
}

func (m *minioService) buildAccessURL(bucket string, key string) string {
	return strings.Join([]string{m.minioClient.EndpointURL().String(), bucket, key}, "/")
}

func NewMinioService(conf *viper.Viper) Service {
	minioClient, err := minio.New(conf.GetString("app.data.oss.endpoint"), &minio.Options{
		Creds:  credentials.NewStaticV4(conf.GetString("app.data.oss.access_key"), conf.GetString("app.oss.secret_key"), conf.GetString("app.oss.token")),
		Secure: conf.GetBool("app.data.oss.use_ssl"),
	})
	if err != nil {
		panic(err)
	}
	return &minioService{
		minioClient: minioClient,
		expires:     conf.GetInt64("app.data.oss.expires"),
	}
}
