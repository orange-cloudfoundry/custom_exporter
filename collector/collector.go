package collector

import (
	"errors"
	"fmt"
	"github.com/orange-cloudfoundry/custom_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"strings"
	"time"
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

// Exporter collects MySQL metrics. It implements prometheus.Collector.
type CollectorHelper struct {
	duration, error prometheus.Gauge
	totalScrapes    prometheus.Counter
	scrapeErrors    *prometheus.CounterVec
	collectorCustom CollectorCustom
}
type CollectorCustom interface {
	Name() string
	Desc() string
	Run(ch chan<- prometheus.Metric) error
	Config() config.MetricsItem
}

func NewCollectorHelper(collectorCustom CollectorCustom) *CollectorHelper {
	configName := collectorCustom.Config().Name

	helper := &CollectorHelper{
		duration: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: config.Namespace,
			Subsystem: configName,
			Name:      "last_scrape_duration_seconds",
			Help:      "Duration of the last scrape of metrics from " + configName,
		}),

		error: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: config.Namespace,
			Subsystem: configName,
			Name:      "last_scrape_error",
			Help:      "Whether the last scrape of metrics from " + configName + " resulted in an error (1 for error, 0 for success).",
		}),

		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: config.Namespace,
			Subsystem: configName,
			Name:      "scrapes_total",
			Help:      "Total number of times " + configName + " was scraped for metrics.",
		}),

		scrapeErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: config.Namespace,
			Subsystem: configName,
			Name:      "scrape_errors_total",
			Help:      "Total number of times an error occurred scraping a " + configName,
		}, []string{"collector"}),

		collectorCustom: collectorCustom,
	}

	return helper
}

func (e CollectorHelper) Check(err error) error {
	config := e.collectorCustom.Config()
	name := e.collectorCustom.Name()

	if config.Credential.Collector != name {
		err = errors.New(
			fmt.Sprintf("Error mismatching collector type : config type = %s & current type = %s",
				config.Credential.Collector,
				name,
			))
		log.Errorln("Error:", err)
	}

	if len(config.Commands) < 1 {
		err = errors.New("Error empty commands to run !!")
		log.Errorln("Error:", err)
	}

	return err
}

func (e *CollectorHelper) Describe(ch chan<- *prometheus.Desc) {
	log.Debugln("Call Shell Describe")

	metricCh := make(chan prometheus.Metric)
	doneCh := make(chan struct{})

	go func() {
		for m := range metricCh {
			ch <- m.Desc()
		}
		close(doneCh)
	}()

	e.Collect(metricCh)
	close(metricCh)
	<-doneCh
}

// Collect implements prometheus.Collector.
func (e *CollectorHelper) Collect(ch chan<- prometheus.Metric) {
	log.Debugln("Call Generic Collect")
	e.scrape(ch)
	ch <- e.duration
	ch <- e.totalScrapes
	ch <- e.error
	e.scrapeErrors.Collect(ch)
}

func (e *CollectorHelper) scrape(ch chan<- prometheus.Metric) {
	log.Debugln("Call Shell scrape")
	e.totalScrapes.Inc()

	var err error

	defer func(begun time.Time) {
		e.duration.Set(time.Since(begun).Seconds())
		if err == nil {
			e.error.Set(0)
		} else {
			e.error.Set(1)
		}
	}(time.Now())

	err = e.collectorCustom.Run(ch)
}

func PromDesc(collectorCustom CollectorCustom) string {
	log.Debugln("Call Generic PromDesc")

	var namespace string
	var subsystem string
	var name string

	namespace = config.Namespace
	//subsystem = collectorCustom.Name()
	name = strings.ToLower(collectorCustom.Config().Name)

	log.Debugf("Calling PromDesc with namespace \"%s\", subsystem \"%s\" and name \"%s\"", namespace, subsystem, name)
	return prometheus.BuildFQName(namespace, subsystem, name)
}
