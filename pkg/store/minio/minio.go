package minio

import (
	"context"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/store"

	"github.com/minio/minio-go/v7/pkg/credentials"

	minio "github.com/minio/minio-go/v7"
)

// Config minio client config
type Config struct {
	// object store internal endpoint
	InternalEndpoint string
	// object get external endpoint.
	// it may be a domain or ip+port address
	ExternalEndpoint string
	// bucket region, default empty string
	Region string
	// minio access key
	AccessKey string
	// minio secret key
	SecretKey string
	// http or https, default http
	Secure bool
	// file url expire time. Remember that expired time can't greater than 7 days
	ExpireTime time.Duration
}

// MinioClient minio client struct
type MinioClient struct {
	client *minio.Client
	config *Config
}

// New returns MinioClient instance, it default use S3 V2 signature
// because it will be override if endpoint is S3 or GCS schema
func New(cfg *Config) (clnt *MinioClient, err error) {
	client, err := minio.New(cfg.InternalEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV2(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.Secure,
	})
	if err != nil {
		return
	}

	if cfg.ExternalEndpoint == "" {
		cfg.ExternalEndpoint = cfg.InternalEndpoint
	}

	if cfg.ExpireTime <= 0 {
		cfg.ExpireTime = store.DefaultExpireTime
	}

	clnt = &MinioClient{
		client: client,
		config: cfg,
	}

	return
}

// PutObject creates an object in a bucket.
//
// You must have WRITE permissions on a bucket to create an object.
//
//  - For size smaller than 128MiB PutObject automatically does a
//    single atomic Put operation.
//  - For size larger than 128MiB PutObject automatically does a
//    multipart Put operation.
//  - For size input as -1 PutObject does a multipart Put operation
//    until input stream reaches EOF. Maximum object size that can
//    be uploaded through this operation will be 5TiB.
func (mc *MinioClient) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, ext ...string) (fileUrl string, err error) {
	po := minio.PutObjectOptions{}
	if len(ext) > 0 {
		po.ContentType = store.TypeByExtension(ext[0])
	}

	if _, err = mc.client.PutObject(ctx, bucketName, objectName, reader, objectSize, minio.PutObjectOptions{}); err != nil {
		return
	}

	return mc.GetFileUrl(ctx, bucketName, objectName)
}

// FPutObject create an object in a bucket, with contents from file at filePath
func (mc *MinioClient) FPutObject(ctx context.Context, bucketName, objectName, path string, ext ...string) (fileUrl string, err error) {
	po := minio.PutObjectOptions{}
	if len(ext) > 0 {
		po.ContentType = store.TypeByExtension(ext[0])
	}

	if _, err = mc.client.FPutObject(ctx, bucketName, objectName, path, po); err != nil {
		return
	}

	return mc.GetFileUrl(ctx, bucketName, objectName)
}

// GetFileUrl generate presigned get object url
func (mc *MinioClient) GetFileUrl(ctx context.Context, bucketName string, objectName string, expireTime ...int64) (fileUrl string, err error) {
	expired := mc.config.ExpireTime
	if len(expireTime) > 0 {
		expired = time.Duration(expireTime[0]) * time.Second
	}
	// Set request parameters
	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", "attachment; filename="+objectName)
	presignedURL, err := mc.client.PresignedGetObject(ctx, bucketName, objectName, expired, reqParams)
	fileUrl = presignedURL.String()
	if mc.config.InternalEndpoint != mc.config.ExternalEndpoint {
		fileUrl = strings.Replace(fileUrl, mc.config.InternalEndpoint, mc.config.ExternalEndpoint, 1)
	}

	return
}
