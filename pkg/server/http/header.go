package http

import (
	"net/http"

	"github.com/UnderTreeTech/waterdrop/pkg/utils/xstring"

	"github.com/UnderTreeTech/waterdrop/pkg/status"
	"github.com/UnderTreeTech/waterdrop/pkg/utils/xreply"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/gin-gonic/gin"
)

// Header middleware is commonly used for p2p communication, like ios/android application, or server to server call
func (s *Server) Header() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		if c.Request.Method != http.MethodGet {
			if "application/json" != xstring.StripContentType(c.Request.Header.Get(_contentType)) {
				log.Warn(ctx, "invalid content-type", log.String("content-type", c.Request.Header.Get(_contentType)))
				c.AbortWithStatusJSON(http.StatusBadRequest, xreply.Reply(ctx, nil, status.RequestErr))
				return
			}
		}

		appkey := c.Request.Header.Get(_appkey)
		if _appkeyLen != len(appkey) {
			log.Warn(ctx, "fake appkey", log.String("appkey", appkey))
			c.AbortWithStatusJSON(http.StatusBadRequest, xreply.Reply(ctx, nil, status.AppKeyInvalid))
		}

		c.Next()
	}
}
