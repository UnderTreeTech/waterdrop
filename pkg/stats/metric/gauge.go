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

type GaugeVecOpts struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
	Labels    []string
}

type gaugeVec struct {
	*prometheus.GaugeVec
}

func NewGaugeVec(opt *GaugeVecOpts) *gaugeVec {
	vector := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: opt.Namespace,
			Subsystem: opt.Subsystem,
			Name:      opt.Name,
			Help:      opt.Help,
		}, opt.Labels)
	prometheus.MustRegister(vector)

	return &gaugeVec{GaugeVec: vector}
}

func (g *gaugeVec) Inc(labels ...string) {
	g.WithLabelValues(labels...).Inc()
}

func (g *gaugeVec) Dec(labels ...string) {
	g.WithLabelValues(labels...).Dec()
}

func (g *gaugeVec) Add(v float64, labels ...string) {
	g.WithLabelValues(labels...).Add(v)
}

func (g *gaugeVec) Sub(v float64, labels ...string) {
	g.WithLabelValues(labels...).Sub(v)
}
