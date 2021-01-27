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
	"fmt"
	"net/http"

	es7 "github.com/olivere/elastic/v7"
	es5 "gopkg.in/olivere/elastic.v5"
	es6 "gopkg.in/olivere/elastic.v6"
)

type Config struct {
	Username string
	Password string
	Version  string
	Schema   string // default http
	URLs     []string
	Plugins  []string
}

type ClientV5 struct {
	*es5.Client
	*Config
}

type ClientV6 struct {
	*es6.Client
	*Config
}

type ClientV7 struct {
	*es7.Client
	*Config
}

// NewClient is a client factory, it returns client according with version
func NewClient(config *Config) interface{} {
	if "" == config.Version {
		panic(fmt.Sprintf("Must config es version, Currently version only can be v5, v6 or v7."))
	}

	if "" == config.Schema {
		config.Schema = "http"
	}

	switch config.Version {
	case "v5":
		return clientV5(config)
	case "v6":
		return clientV6(config)
	case "v7":
		return clientV7(config)
	default:
		panic(fmt.Sprintf("Unsupport es version %s.Currently it must be configed as v5, v6 or v7 ", config.Version))
	}

	return nil
}

func clientV5(config *Config) *ClientV5 {
	es5, _ := es5.NewClient(
		es5.SetHttpClient(&http.Client{
			Transport: NewTransport(config),
		}),
		es5.SetBasicAuth(config.Username, config.Password),
		es5.SetURL(config.URLs...),
		es5.SetScheme(config.Schema),
		es5.SetRequiredPlugins(config.Plugins...),
	)

	client := &ClientV5{
		Config: config,
		Client: es5,
	}
	return client
}

func clientV6(config *Config) *ClientV6 {
	es6, _ := es6.NewClient(
		es6.SetHttpClient(&http.Client{
			Transport: NewTransport(config),
		}),
		es6.SetBasicAuth(config.Username, config.Password),
		es6.SetURL(config.URLs...),
		es6.SetScheme(config.Schema),
		es6.SetRequiredPlugins(config.Plugins...),
	)
	client := &ClientV6{
		Config: config,
		Client: es6,
	}
	return client
}

func clientV7(config *Config) *ClientV7 {
	es7, _ := es7.NewClient(
		es7.SetHttpClient(&http.Client{
			Transport: NewTransport(config),
		}),
		es7.SetBasicAuth(config.Username, config.Password),
		es7.SetURL(config.URLs...),
		es7.SetScheme(config.Schema),
		es7.SetRequiredPlugins(config.Plugins...),
	)
	client := &ClientV7{
		Config: config,
		Client: es7,
	}
	return client
}
