package minio

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFileUrl(t *testing.T) {
	cfg := &Config{
		InternalEndpoint: "localhost:9000",
		ExternalEndpoint: "localhost:8000",
		Region:           "",
		AccessKey:        "H1ZRYMN0LG5T7TP3AYQX",
		SecretKey:        "So6itWRcVJVI8ySXfBSOf+p72+dQeDbxk9otkyR8",
	}
	clnt, err := New(cfg)
	assert.Nil(t, err)
	path, err := clnt.GetFileUrl("john", "##blog-master+#.tar.gz")
	assert.Nil(t, err)
	fmt.Println(path)
}
