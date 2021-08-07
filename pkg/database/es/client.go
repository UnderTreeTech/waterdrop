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

package es

import (
	"net/http"

	es7 "github.com/olivere/elastic/v7"
)

type Config struct {
	Username string
	Password string
	Version  string
	Schema   string // default http
	URLs     []string
	Plugins  []string
}

type Client struct {
	*es7.Client
	*Config
}

// NewClient returns es7 client pointer
func NewClient(config *Config) *Client {
	if "" == config.Schema {
		config.Schema = "http"
	}

	es7, _ := es7.NewClient(
		es7.SetHttpClient(&http.Client{
			Transport: NewTransport(config),
		}),
		es7.SetBasicAuth(config.Username, config.Password),
		es7.SetURL(config.URLs...),
		es7.SetScheme(config.Schema),
		es7.SetRequiredPlugins(config.Plugins...),
	)
	client := &Client{
		Config: config,
		Client: es7,
	}
	return client
}
