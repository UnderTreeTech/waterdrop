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

package utils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

var envCache struct {
	once sync.Once
	m    map[string]string
}

// EnvFile returns the name of the Go environment configuration file.
func EnvFile() (string, error) {
	if file := os.Getenv("GOENV"); file != "" {
		if file == "off" {
			return "", fmt.Errorf("GOENV=off")
		}
		return file, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	if dir == "" {
		return "", fmt.Errorf("missing user-config dir")
	}
	return filepath.Join(dir, "go/env"), nil
}

// initEnvCache init environment cahce
func initEnvCache() {
	envCache.m = make(map[string]string)
	file, _ := EnvFile()
	if file == "" {
		return
	}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}

	for len(data) > 0 {
		// Get next line.
		line := data
		i := bytes.IndexByte(data, '\n')
		if i >= 0 {
			line, data = line[:i], data[i+1:]
		} else {
			data = nil
		}

		i = bytes.IndexByte(line, '=')
		if i < 0 || line[0] < 'A' || 'Z' < line[0] {
			// Line is missing = (or empty) or a comment or not a valid env name. Ignore.
			// (This should not happen, since the file should be maintained almost
			// exclusively by "go env -w", but better to silently ignore than to make
			// the go command unusable just because somehow the env file has
			// gotten corrupted.)
			continue
		}
		key, val := line[:i], line[i+1:]
		envCache.m[string(key)] = string(val)
	}
}

// Getenv gets the value from env or configuration.
func Getenv(key string) string {
	val := os.Getenv(key)
	if val != "" {
		return val
	}
	envCache.once.Do(initEnvCache)
	return envCache.m[key]
}
