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
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/alibaba/sentinel-golang/api"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/util"

	"github.com/alibaba/sentinel-golang/core/flow"
)

func TestSentinel(t *testing.T) {
	config := &Config{
		AppName:   "sentinel",
		FlowRules: make([]*flow.Rule, 0),
	}

	config.FlowRules = append(config.FlowRules, &flow.Rule{
		Resource:               "sentinel",
		TokenCalculateStrategy: flow.Direct,
		ControlBehavior:        flow.Reject,
		Threshold:              10,
		StatIntervalInMs:       1000,
	})

	InitSentinel(config)

	for i := 0; i < 10; i++ {
		go func() {
			for {
				e, b := api.Entry("sentinel", api.WithTrafficType(base.Inbound))
				if b != nil {
					// Blocked. We could get the block reason from the BlockError.
					fmt.Println("blocked request", b.BlockMsg())
					time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)
				} else {
					// Passed, wrap the logic here.
					fmt.Println(util.CurrentTimeMillis(), "passed")
					time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)

					// Be sure the entry is exited finally.
					e.Exit()
				}
			}
		}()
	}

	time.Sleep(time.Second * 1)
}
