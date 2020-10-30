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

package setinel

import (
	"encoding/json"
	"io/ioutil"

	"github.com/alibaba/sentinel-golang/api"

	sc "github.com/alibaba/sentinel-golang/core/config"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/alibaba/sentinel-golang/core/flow"
)

type Config struct {
	AppName   string
	LogPath   string
	FlowRules []*flow.Rule
	RulePath  string
}

func InitSentinel(config *Config) error {
	if config.RulePath != "" {
		var rules []*flow.Rule
		content, err := ioutil.ReadFile(config.RulePath)
		if err != nil {
			log.Errorf("read rule fail", log.String("rule_path", config.RulePath), log.String("error", err.Error()))
		}

		if err := json.Unmarshal(content, &rules); err != nil {
			log.Errorf("unmarshal rule fail", log.String("rule_path", config.RulePath), log.String("error", err.Error()))
		}

		config.FlowRules = append(config.FlowRules, rules...)
	}

	entity := sc.NewDefaultConfig()
	entity.Sentinel.App.Name = config.AppName
	entity.Sentinel.Log.Dir = config.LogPath

	if len(config.FlowRules) > 0 {
		if _, err := flow.LoadRules(config.FlowRules); err != nil {
			log.Errorf("load rule fail", log.String("rule_path", config.RulePath), log.String("error", err.Error()))
		}
	}

	return api.InitWithConfig(entity)
}
