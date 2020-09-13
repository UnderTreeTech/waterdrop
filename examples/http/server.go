package main

import (
	"context"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http"
	"github.com/gin-gonic/gin"
)

func main() {
	defer log.New(nil).Sync()

	srv := http.NewServer(nil)

	g := srv.Group("/api")
	{
		g.GET("/ping", ping)
		g.GET("/waterdrop", waterdrop)
	}

	srv.Start()

	time.Sleep(time.Minute * 5)
	srv.Stop(context.Background())
}

func ping(c *gin.Context) {
	c.JSON(200, "ping")
}

func waterdrop(c *gin.Context) {
	c.JSON(200, "Framwork waterdrop")
}
