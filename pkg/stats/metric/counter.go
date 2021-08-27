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

package metric

import "github.com/prometheus/client_golang/prometheus"

// CounterVecOpts
type CounterVecOpts struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
	Labels    []string
}

type counterVec struct {
	*prometheus.CounterVec
}

// NewCounterVec returns a CounterVecOpts instance
func NewCounterVec(opt *CounterVecOpts) *counterVec {
	vector := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: opt.Namespace,
			Subsystem: opt.Subsystem,
			Name:      opt.Name,
			Help:      opt.Help,
		}, opt.Labels)
	prometheus.MustRegister(vector)

	return &counterVec{CounterVec: vector}
}

// Inc increments the counter by 1
func (c *counterVec) Inc(labels ...string) {
	c.WithLabelValues(labels...).Inc()
}

// Add adds the given value to the counter. It panics if the value is < 0
func (c *counterVec) Add(v float64, labels ...string) {
	c.WithLabelValues(labels...).Add(v)
}
