package collector

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/orange-cloudfoundry/custom_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"gopkg.in/redis.v5"
)

/*
Copyright 2017 Orange

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

const (
	CollectorRedisName = "redis"
	CollectorRedisDesc = "Metrics from redis collector in the custom exporter."
)

type CollectorRedis struct {
	metricsConfig config.MetricsItem
}

func NewCollectorRedis(config config.MetricsItem) *CollectorRedis {
	return &CollectorRedis{
		metricsConfig: config,
	}
}

func NewPrometheusRedisCollector(config config.MetricsItem) (prometheus.Collector, error) {
	var err error

	myCol := NewCollectorHelper(NewCollectorRedis(config))

	log.Infof("Collector Added: Type '%s' / Name '%s' / Credentials '%s'", CollectorRedisName, config.Name, config.Credential.Name)

	if len(config.Value_name) < 1 {
		err = fmt.Errorf("keymapping not present for collector %s", CollectorRedisName)
		log.Errorln("Error:", err)
	}

	return myCol, myCol.Check(err)
}

func (e CollectorRedis) Config() config.MetricsItem {
	return e.metricsConfig
}

func (e CollectorRedis) Name() string {
	return CollectorRedisName
}

func (e CollectorRedis) Desc() string {
	return CollectorRedisDesc
}

func (e *CollectorRedis) Run(ch chan<- prometheus.Metric) error {
	var (
		red      *redis.Client
		jsn      map[string]interface{}
		res      map[string]string
		labelVal []string
		mapping  []string
		err      error
		out      []byte
	)

	err = nil
	mapping = e.metricsConfig.Mapping
	labelVal = make([]string, len(mapping))

	if red, err = e.redisClient(); err != nil {
		log.Errorf("Error when get Redis Client for metric \"%s\" : %s", e.metricsConfig.Name, err.Error())
		return err
	}

	defer red.Close()

	log.Debugln("Calling Redis Commands... ")

	for _, c := range e.metricsConfig.Commands {
		c = strings.TrimSpace(c)

		if len(c) < 1 {
			continue
		}

		cmd := e.redisRun(red, c)

		if cmd.Err() != nil {
			log.Errorf("Error for metrics \"%s\" while running redis command \"%s\": %s", e.metricsConfig.Name, c, cmd.Err().Error())
			return cmd.Err()
		}

		out = []byte(cmd.Val().(string))
		jsn = make(map[string]interface{})

		if err = json.Unmarshal(out, &jsn); err != nil {
			log.Errorf("Error for metrics \"%s\" while parsing json result of redis command \"%s\": %s", e.metricsConfig.Name, c, err.Error())
			return err
		}
	}

	res = make(map[string]string)
	for k, v := range jsn {
		res[k] = e.interface2String(v)
	}

	log.Debugln("Filtering Redis Label Value... ")

	for i, k := range mapping {
		if val, isOk := res[k]; !isOk {
			log.Debugln("TagValue not found :", k)
			labelVal[i] = ""
		} else {
			labelVal[i] = val
		}
	}

	log.Debugln("Filtering Redis Metric Value... ")
	metricVal := float64(0)

	if val, isOk := res[e.metricsConfig.Value_name]; !isOk {
		err = fmt.Errorf("keymapping not found in resultSet for collector %s and command [ %s ]", CollectorRedisName, strings.Join(e.metricsConfig.Commands, ", "))
	} else {
		if metricVal, err = strconv.ParseFloat(val, 64); err != nil {
			metricVal = float64(0)
		}
	}

	prom_desc := PromDesc(e)
	log.Debugf("Add Metric \"%s\" : Tag '%s' / TagValue '%s' / Value '%v'", prom_desc, mapping, labelVal, metricVal)

	metric := prometheus.MustNewConstMetric(
		prometheus.NewDesc(prom_desc, e.metricsConfig.Name, mapping, nil),
		e.metricsConfig.Value_type, metricVal, labelVal...,
	)

	select {
	case ch <- metric:
		log.Debug("Return no error...")
		return nil
	default:
		log.Info("Cannot write to channel...")
	}

	return err
}

func (e CollectorRedis) interface2String(input interface{}) string {

	if val, ok := input.(float64); ok {
		return strconv.FormatFloat(val, 'f', -1, 64)
	}

	if val, ok := input.(float32); ok {
		return strconv.FormatFloat(float64(val), 'f', -1, 32)
	}

	if val, ok := input.(int); ok {
		return strconv.FormatInt(int64(val), 10)
	}

	if val, ok := input.(bool); ok {
		return strconv.FormatBool(val)
	}

	if val, ok := input.(string); ok {
		return string(val)
	}

	return ""
}

func (e CollectorRedis) DsnPart() (map[string]interface{}, error) {
	var (
		dbn  int64
		dsn  *url.URL
		pss  string
		err  error
		isOk bool
		res  map[string]interface{}
	)

	res = make(map[string]interface{})

	if dsn, err = url.Parse(e.metricsConfig.Credential.Dsn); err != nil {
		return res, err
	}

	if pss, isOk = dsn.User.Password(); !isOk {
		pss = ""
	}

	if strings.Trim(dsn.Path, "/") == "" {
		dbn = 0
	} else {
		if dbn, err = strconv.ParseInt(strings.Trim(dsn.Path, "/"), 10, 64); err != nil {
			return res, fmt.Errorf("db identifier not well formatted (int value required) : %s", err.Error())
		}
	}

	res["addr"] = string(dsn.Host)
	res["pass"] = string(pss)
	res["dbnum"] = int(dbn)

	return res, nil
}

func (e CollectorRedis) redisClient() (*redis.Client, error) {
	var (
		clt *redis.Client
		dsn map[string]interface{}
		err error
	)

	if dsn, err = e.DsnPart(); err != nil {
		return clt, err
	}

	var redisOpt = redis.Options{
		Addr: dsn["addr"].(string),
	}

	if _, ok := dsn["pass"]; ok && len(strings.TrimSpace(dsn["pass"].(string))) > 0 {
		redisOpt.Password = strings.TrimSpace(dsn["pass"].(string))
	}

	if _, ok := dsn["dbnum"]; ok {
		redisOpt.DB = dsn["dbnum"].(int)
	}

	if redisOpt.Password == "" && redisOpt.DB == 0 {
		redisOpt.ReadOnly = true
	}

	log.Debugf("Starting client redis for metrics \"%s\", with params : %v", e.metricsConfig.Name, redisOpt)
	clt = redis.NewClient(&redisOpt)

	return clt, e.redisPing(clt)
}

func (e CollectorRedis) redisPing(client *redis.Client) error {
	if _, err := client.Ping().Result(); err != nil {
		return err
	}

	return nil
}

func (e CollectorRedis) redisRun(client *redis.Client, command string) *redis.Cmd {
	var (
		arg []interface{}
		res *redis.Cmd
	)

	cmd := strings.Split(command, " ")
	arg = make([]interface{}, len(cmd))

	for k, v := range cmd {
		arg[k] = v
	}

	log.Debugf("Prepare command for metrics \"%s\" : %v", e.metricsConfig.Name, arg)
	res = redis.NewCmd(arg...)

	if res.Err() != nil {
		log.Errorf("Error with metrics \"%s\" for command \"%s\" : %s", e.metricsConfig.Name, command, res.Err().Error())
		return res
	}

	log.Debugf("Proceed command for metrics \"%s\"...", e.metricsConfig.Name)

	client.Process(res)
	log.Debugf("Proceded command for metrics \"%s\" : %v", e.metricsConfig.Name, res)

	return res
}
