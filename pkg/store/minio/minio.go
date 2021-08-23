package minio

import (
	"context"
	"io"
	"net/url"
	"time"

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
		cfg.ExpireTime = time.Second * 5 * 60
	}

	clnt = &MinioClient{
		client: client,
		config: cfg,
	}
	return
}

func (mc *MinioClient) PutObject(bucketName string, objectName string, reader io.Reader, objectSize int64) {
	mc.client.PutObject(context.Background(), bucketName, objectName, reader, objectSize, minio.PutObjectOptions{})
}

func (mc *MinioClient) UploadFile(bucketName string, objectName string, file string) (path string, err error) {
	return
}
func (mc *MinioClient) DownloadFile(bucketName string, objectName string) (file []byte, err error) {
	return
}

// GetFileUrl generate presigned get object url
func (mc *MinioClient) GetFileUrl(bucketName string, objectName string, expireTime ...int64) (path string, err error) {
	expired := mc.config.ExpireTime
	if len(expireTime) > 0 {
		expired = time.Duration(expireTime[0]) * time.Second
	}
	// Set request parameters
	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", "attachment; filename="+objectName)
	presignedURL, err := mc.client.PresignedGetObject(context.Background(), bucketName, objectName, expired, reqParams)
	if mc.config.InternalEndpoint != mc.config.ExternalEndpoint {
		presignedURL.Host = mc.config.ExternalEndpoint
	}

	path = presignedURL.String()
	return
}
