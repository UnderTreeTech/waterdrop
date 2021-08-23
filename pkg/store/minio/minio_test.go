package minio

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFileUrl(t *testing.T) {
	cfg := &Config{
		InternalEndpoint: "localhost:9000",
		ExternalEndpoint: "localhost:9000",
		Region:           "",
		AccessKey:        "XMFMOKB2FJWA0I9JIR62",
		SecretKey:        "aMRKmxoRcb+Ezr5CmOmFAqFwYWPrEFA7UdtWWWOl",
	}
	clnt, err := New(cfg)
	assert.Nil(t, err)
	ctx := context.Background()

	exsit, err := clnt.ExistBucket(ctx, "test")
	assert.Nil(t, err)
	if !exsit {
		err = clnt.CreateBucket(ctx, "test")
		assert.Nil(t, err)
	}

	_, err = clnt.FPutObject(ctx, "test", "minio.go", "./minio.go")
	assert.Nil(t, err)

	object, err := os.Open("./minio_test.go")
	assert.Nil(t, err)
	defer object.Close()

	stat, err := object.Stat()
	assert.Nil(t, err)
	_, err = clnt.PutObject(ctx, "test", "minio_test.go", object, stat.Size(), ".jpeg")
	assert.Nil(t, err)
	path, err := clnt.GetFileUrl(ctx, "test", "minio_test.go")
	assert.Nil(t, err)
	assert.Contains(t, path, "minio_test.go")
}
