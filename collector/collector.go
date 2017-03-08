package collector

import (
	"github.com/orange-cloudfoundry/custom_exporter/custom_config"
	"github.com/prometheus/client_golang/prometheus"
	"strings"
	"github.com/prometheus/common/log"
	"time"
	"fmt"
	"errors"
)

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
	Run(ch chan <- prometheus.Metric) error
	Config() custom_config.MetricsItem
}

func NewCollectorHelper(collectorCustom CollectorCustom) *CollectorHelper {
	collectorName := collectorCustom.Name()
	configName := collectorCustom.Config().Name

	if len(collectorName) < 1 {
		collectorName = custom_config.Exporter
	}

	helper := &CollectorHelper{
		duration: prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: custom_config.Namespace,
		Subsystem: collectorName,
		Name:      "last_scrape_duration_seconds",
		Help:      "Duration of the last scrape of metrics from " + configName,
		}),

		error: prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: custom_config.Namespace,
		Subsystem: collectorName,
		Name:      "last_scrape_error",
		Help:      "Whether the last scrape of metrics from " + configName + " resulted in an error (1 for error, 0 for success).",
		}),

		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: custom_config.Namespace,
		Subsystem: collectorName,
		Name:      "scrapes_total",
		Help:      "Total number of times " + configName + " was scraped for metrics.",
		}),

		scrapeErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: custom_config.Namespace,
		Subsystem: collectorName,
		Name:      "scrape_errors_total",
		Help:      "Total number of times an error occurred scraping a " + configName,
		}, []string{"collector"}),

		collectorCustom: collectorCustom,
	}

	return helper
}

func (e CollectorHelper) Check() error {
	config := e.collectorCustom.Config()
	name := e.collectorCustom.Name()

	if config.Credential.Collector != name {
		err := errors.New(
			fmt.Sprintf("Error mismatching collector type : config type = %s & current type = %s",
				config.Credential.Collector,
				name,
			))
		log.Errorln("Error:", err)
		return err
	}


	if len(config.Commands) < 1 {
		err := errors.New("Error empty commands to run !!")
		log.Errorln("Error:", err)
		return err
	}

	return nil
}

func (e *CollectorHelper) Describe(ch chan <- *prometheus.Desc) {
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
func (e *CollectorHelper) Collect(ch chan <- prometheus.Metric) {
	log.Debugln("Call Generic Collect")
	e.scrape(ch)
	ch <- e.duration
	ch <- e.totalScrapes
	ch <- e.error
	e.scrapeErrors.Collect(ch)
}

func (e *CollectorHelper) scrape(ch chan <- prometheus.Metric) {
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
	return prometheus.BuildFQName(
		custom_config.Namespace,
		collectorCustom.Name(),
		strings.ToLower(collectorCustom.Config().Name),
	)
}
