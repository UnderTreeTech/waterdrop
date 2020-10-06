package http

import (
	"net/http"
	"strconv"

	"github.com/UnderTreeTech/waterdrop/pkg/utils/xstring"

	"github.com/UnderTreeTech/waterdrop/pkg/status"
	"github.com/UnderTreeTech/waterdrop/pkg/utils/xreply"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/gin-gonic/gin"
)

func (s *Server) Header() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		if c.Request.Method != http.MethodGet {
			if "application/json" != xstring.StripContentType(c.Request.Header.Get(_contentType)) {
				log.Warn(ctx, "invalid content-type", log.String("content-type", c.Request.Header.Get(_contentType)))
				c.Error(status.RequestErr)
				c.AbortWithStatusJSON(http.StatusBadRequest, xreply.Reply(nil, status.RequestErr))
				return
			}
		}

		_, err := strconv.Atoi(c.Request.Header.Get("timestamp"))
		if err != nil {
			log.Warn(ctx, "invalid timestamp", log.String("timestamp", c.Request.Header.Get("timestamp")))
			c.Error(status.RequestErr)
			c.AbortWithStatusJSON(http.StatusBadRequest, xreply.Reply(nil, status.RequestErr))
			return
		}

		appkey := c.Request.Header.Get(_appkey)
		if _appkeyLen != len(appkey) {
			log.Warn(ctx, "fake appkey", log.String("appkey", appkey))
			c.Error(status.AppKeyInvalid)
			c.AbortWithStatusJSON(http.StatusBadRequest, xreply.Reply(nil, status.AppKeyInvalid))
		}

		c.Next()
	}
}
