// Copyright 2022 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package metrics

import (
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

type MemoryCollector struct {
	client       *Client
	podName      string
	instanceName string
	namespace    string
	resourceId   string
	logger       log.Logger
}

var (
	freeMemoryDesc = prometheus.NewDesc(
		prometheus.BuildFQName("aws", "rds", "freeable_memory_average"),
		"postgres_exporter: Freeable memory",
		[]string{"dbinstance_identifier", "exported_job", "instance", "job"},
		nil,
	)
)

func NewMemoryUsageCollector(client *Client, podName, instanceName, namespace, resourceId string, logger log.Logger) *MemoryCollector {
	return &MemoryCollector{client: client, podName: podName, instanceName: instanceName, namespace: namespace, resourceId: resourceId, logger: logger}
}

func (m *MemoryCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- freeMemoryDesc
}

func (m *MemoryCollector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		query := fmt.Sprintf("query=(container_memory_max_usage_bytes{container=\"pg-%s\", namespace=\"%s\", pod=\"%s\"} - container_memory_usage_bytes{container=\"pg-%s\",namespace=\"%s\",pod=\"%s\"})", m.resourceId, m.namespace, m.podName, m.resourceId, m.namespace, m.podName)
		resp, err := m.client.execute("GET", "/api/v1/query", query)
		if err != nil {
			level.Debug(m.logger).Log("msg", "collector returned no data", "err", err)
			return
		}
		if len(resp.Data.Result) == 0 {
			level.Debug(m.logger).Log("msg", "collector returned no data", "err", err)
			return
		}
		val, err := getFloat(resp.Data.Result[0].Value[1])
		if err != nil {
			level.Debug(m.logger).Log("msg", "collector returned no data", "err", err)
			return
		}
		fmt.Println(val)
		ch <- prometheus.MustNewConstMetric(freeMemoryDesc,
			prometheus.GaugeValue,
			val,
			m.instanceName,
			"aws_rds",
			resp.Data.Result[0].Metric.Instance,
			"rds",
		)
	}()
	wg.Wait()
}
