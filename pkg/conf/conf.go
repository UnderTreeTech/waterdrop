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

package conf

import (
	"flag"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/conf/parser/toml"

	"github.com/UnderTreeTech/waterdrop/pkg/conf/provider/file"

	"github.com/mitchellh/mapstructure"
)

var (
	defaultDelimiter = "."
	defaultConfTag   = "conf"

	confPath    string
	watchConfig bool

	defaultConfig *Config
)

type Provider interface {
	ReadBytes() ([]byte, error)
	Watch(func()) error
}

type Parser interface {
	Marshal(map[string]interface{}) ([]byte, error)
	Unmarshal([]byte) (map[string]interface{}, error)
}

func init() {
	flag.StringVar(&confPath, "conf", "", "default config path")
	flag.BoolVar(&watchConfig, "watch", false, "default watch config param")
}

func Init() {
	if confPath != "" {
		defaultConfig = New()
		provider := file.NewFileProvider(confPath, watchConfig)
		parser := toml.NewTOMLParser()
		if err := defaultConfig.Load(provider, parser); err != nil {
			panic(fmt.Sprintf("load config fail,err msg %s", err.Error()))
		}

		if provider.IsEnableWatch() {
			provider.Watch(func() {
				time.Sleep(time.Millisecond * 10)
				defaultConfig.Load(provider, toml.NewTOMLParser())
				for _, change := range defaultConfig.onChanges {
					change(defaultConfig)
				}
			})
		}
	} else {
		panic("only support file config now,lack of remote config center args")
	}
}

func Unmarshal(key string, object interface{}) error {
	return defaultConfig.Unmarshal(key, object)
}

func KeyMap() map[string]interface{} {
	return defaultConfig.GetKeyMap()
}

func OnChange(cb func(*Config)) {
	defaultConfig.OnChange(cb)
}

type Config struct {
	mutex     sync.RWMutex
	keyMap    map[string]interface{}
	delimiter string

	onChanges []func(*Config)
}

func New() *Config {
	return &Config{
		delimiter: defaultDelimiter,
		keyMap:    make(map[string]interface{}),
		onChanges: make([]func(*Config), 0),
	}
}

func (c *Config) SetDelimiter(delimiter string) {
	c.delimiter = delimiter
}

func (c *Config) GetKeyMap() map[string]interface{} {
	return c.keyMap
}

func (c *Config) OnChange(cb func(*Config)) {
	c.onChanges = append(c.onChanges, cb)
}

// Load takes a Provider that either provides a parsed config map[string]interface{}
// in which case pa (Parser) can be nil, or raw bytes to be parsed, where a Parser
// can be provided to parse.
func (c *Config) Load(provider Provider, parser Parser) error {
	b, err := provider.ReadBytes()
	if err != nil {
		return err
	}

	conf, err := parser.Unmarshal(b)
	if err != nil {
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.keyMap = conf

	return nil
}

// Keys returns the slice of all flattened keys in the loaded configuration
// sorted alphabetically.
func (c *Config) Keys() []string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	keys := make([]string, 0, len(c.keyMap))
	for key := range c.keyMap {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	return keys
}

// Print prints a key -> value string representation
// of the config map with keys sorted alphabetically.
func (c *Config) Print() string {
	sb := strings.Builder{}
	keys := c.Keys()
	for _, key := range keys {
		sb.WriteString(fmt.Sprintf("%s -> %v\n", key, c.keyMap[key]))
	}

	return sb.String()
}

// Unmarshal unmarshals a given key path into the given struct using
// the mapstructure lib. If no path is specified, the whole map is unmarshalled.
// `conf` is the struct field tag used to match field names.
func (c *Config) Unmarshal(key string, object interface{}) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	dc := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.StringToTimeDurationHookFunc(),
		Result:     object,
		TagName:    defaultConfTag,
	}

	decoder, err := mapstructure.NewDecoder(dc)
	if err != nil {
		return err
	}

	m := c.get(key)

	return decoder.Decode(m)
}

// Get returns the raw, uncast interface{} value of a given key path
// in the config map. If the key path does not exist, nil is returned.
func (c *Config) get(key string) interface{} {
	if key == "" {
		return c.keyMap
	}

	val, ok := c.keyMap[key]
	if ok {
		return val
	}

	keys := strings.Split(key, defaultDelimiter)
	res := c.searchKey(c.keyMap, keys)

	return res
}

// Search recursively searches a map for a given path. The path is
// the key map slice, for eg:, parent.child.key -> [parent child key].
//
// It's important to note that all nested maps should be
// map[string]interface{} and not map[interface{}]interface{}.
func (c *Config) searchKey(mp map[string]interface{}, path []string) interface{} {
	next, ok := mp[path[0]]
	if ok {
		if len(path) == 1 {
			return next
		}
		switch next.(type) {
		case map[string]interface{}:
			return c.searchKey(next.(map[string]interface{}), path[1:])
		default:
			return nil
		}
	}
	return nil
}
