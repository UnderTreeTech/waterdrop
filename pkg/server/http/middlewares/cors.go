/*
 *
 * Copyright 2021 waterdrop authors.
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

package middlewares

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSConfig is an alias of cors.Config
type CORSConfig cors.Config

// DefaultCORS default cors handler
func DefaultCORS() gin.HandlerFunc {
	return cors.Default()
}

// NewCORS customer cors handler by config
func NewCORS(config CORSConfig) gin.HandlerFunc {
	return cors.New(cors.Config(config))
}
