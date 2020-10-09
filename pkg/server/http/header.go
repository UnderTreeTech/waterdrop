/*
 *
 * Copyright 2020 waterdrop authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

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
