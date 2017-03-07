package collector

import (
	"github.com/orange-cloudfoundry/custom_exporter/custom_config"
	"github.com/prometheus/client_golang/prometheus"
	"strings"
	"github.com/prometheus/common/log"
	"time"
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
	Run() (string, error)
	Config() custom_config.MetricsItem
}

func NewCollectorHelper(collectorCustom CollectorCustom) *CollectorHelper {
	helper := &CollectorHelper{collectorCustom: collectorCustom}
	helper.setConfig()
	return helper
}
func (c *CollectorHelper) setConfig() {
	collectorName := c.collectorCustom.Name()
	configName := c.collectorCustom.Config().Name
	if len(collectorName) < 1 {
		collectorName = custom_config.Exporter
	}

	c.duration = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: custom_config.Namespace,
		Subsystem: collectorName,
		Name:      "last_scrape_duration_seconds",
		Help:      "Duration of the last scrape of metrics from " + configName,
	})

	c.totalScrapes = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: custom_config.Namespace,
		Subsystem: collectorName,
		Name:      "scrapes_total",
		Help:      "Total number of times " + configName + " was scraped for metrics.",
	})
	c.scrapeErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: custom_config.Namespace,
		Subsystem: collectorName,
		Name:      "scrape_errors_total",
		Help:      "Total number of times an error occurred scraping a " + configName,
	}, []string{"collector"})

	c.error = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: custom_config.Namespace,
		Subsystem: collectorName,
		Name:      "last_scrape_error",
		Help:      "Whether the last scrape of metrics from " + configName + " resulted in an error (1 for error, 0 for success).",
	})
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

	e.collectorCustom.Run()
}
func PromDesc(collectorCustom CollectorCustom) string {
	log.Debugln("Call Generic PromDesc")
	return prometheus.BuildFQName(
		custom_config.Namespace,
		collectorCustom.Name(),
		strings.ToLower(collectorCustom.Config().Name),
	)
}
