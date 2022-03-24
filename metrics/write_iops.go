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

type WriteiopsCollector struct {
	client       *Client
	podName      string
	instanceName string
	namespace    string
	logger       log.Logger
}

var (
	writeIopsDesc = prometheus.NewDesc(
		prometheus.BuildFQName("aws", "rds", "write_iops_average"),
		"postgres_exporter: Write IOPS",
		[]string{"dbinstance_identifier", "exported_job", "instance", "job"},
		nil,
	)
)

func NewWriteiopsCollector(client *Client, podName, instanceName, namespace string, logger log.Logger) *WriteiopsCollector {
	return &WriteiopsCollector{client: client, podName: podName, instanceName: instanceName, namespace: namespace, logger: logger}
}

func (w *WriteiopsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- writeIopsDesc
}

func (w *WriteiopsCollector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		query := fmt.Sprintf("query=sum by (pod, namespace) (rate(container_fs_writes_bytes_total{pod=\"%s\", namespace=\"%s\"}[15m]))", w.podName, w.namespace)
		resp, err := w.client.execute("GET", "/api/v1/query", query)
		if err != nil {
			level.Debug(w.logger).Log("msg", "collector returned no data", "err", err)
			return
		}
		if len(resp.Data.Result) == 0 {
			level.Debug(w.logger).Log("msg", "collector returned no data", "err", err)
			return
		}
		val, err := getFloat(resp.Data.Result[0].Value[1])
		if err != nil {
			level.Debug(w.logger).Log("msg", "collector returned no data", "err", err)
			return
		}
		ch <- prometheus.MustNewConstMetric(writeIopsDesc,
			prometheus.GaugeValue,
			val,
			w.instanceName,
			"aws_rds",
			resp.Data.Result[0].Metric.Instance,
			"rds",
		)
	}()
	wg.Wait()
}
