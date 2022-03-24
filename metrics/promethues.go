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
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"math"
	"reflect"
	"strconv"
)

var floatType = reflect.TypeOf(float64(0))
var stringType = reflect.TypeOf("")

type Client struct {
	host       string
	port       string
	httpClient *resty.Client
}

type Response struct {
	Status string `json:"status,omitempty"`
	Data   Data   `json:"data,omitempty"`
}

type Data struct {
	ResultType string       `json:"resultType,omitempty"`
	Result     []ResultData `json:"result,omitempty"`
}

type ResultData struct {
	Metric MetricData    `json:"metric,omitempty"`
	Value  []interface{} `json:"value,omitempty"`
}

type MetricData struct {
	Name      string `json:"__name__,omitempty"`
	Instance  string `json:"instance,omitempty"`
	Pod       string `json:"pod,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

func NewClient(host, port string) *Client {
	return &Client{host: host, port: port, httpClient: resty.New()}
}

func (c *Client) execute(method, uri, query string) (*Response, error) {
	resp, err := c.httpClient.R().SetHeader("Accept", "application/json").
		SetQueryString(query).
		SetResult(&Response{}).
		Get(fmt.Sprintf("http://%s:%s/%s", c.host, c.port, uri))
	if err != nil && resp == nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, errors.New("not able to get metrics")
	}
	response := resp.Result().(*Response)
	return response, nil
}
func getFloat(unk interface{}) (float64, error) {
	switch i := unk.(type) {
	case float64:
		return i, nil
	case float32:
		return float64(i), nil
	case int64:
		return float64(i), nil
	case int32:
		return float64(i), nil
	case int:
		return float64(i), nil
	case uint64:
		return float64(i), nil
	case uint32:
		return float64(i), nil
	case uint:
		return float64(i), nil
	case string:
		return strconv.ParseFloat(i, 64)
	default:
		v := reflect.ValueOf(unk)
		v = reflect.Indirect(v)
		if v.Type().ConvertibleTo(floatType) {
			fv := v.Convert(floatType)
			return fv.Float(), nil
		} else if v.Type().ConvertibleTo(stringType) {
			sv := v.Convert(stringType)
			s := sv.String()
			return strconv.ParseFloat(s, 64)
		} else {
			return math.NaN(), fmt.Errorf("can't convert %v to float64", v.Type())
		}
	}
}
