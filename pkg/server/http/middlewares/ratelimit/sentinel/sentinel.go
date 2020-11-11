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

package sentinel

import (
	"net/http"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http/middlewares/ratelimit"

	"github.com/UnderTreeTech/waterdrop/pkg/status"

	"github.com/UnderTreeTech/waterdrop/pkg/utils/xreply"

	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/gin-gonic/gin"
)

// Sentinel return rate limit middleware
func Sentinel(opts ...ratelimit.Option) gin.HandlerFunc {
	limitOption := ratelimit.Apply(opts)
	return func(c *gin.Context) {
		limitPath := c.Request.Method + ":" + c.FullPath()
		if limitOption.Strategy != nil {
			limitPath = limitOption.Strategy(c)
		}

		entry, err := api.Entry(limitPath, api.WithResourceType(base.ResTypeWeb), api.WithTrafficType(base.Inbound))
		if err != nil {
			if limitOption.Fallback != nil {
				limitOption.Fallback(c)
			} else {
				c.AbortWithStatusJSON(
					http.StatusTooManyRequests,
					xreply.Reply(c.Request.Context(), nil, status.LimitExceed),
				)
			}
			return
		}

		defer entry.Exit()
		c.Next()
	}
}
