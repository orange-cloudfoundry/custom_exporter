package collector

import (
	"github.com/orange-cloudfoundry/custom_exporter/custom_config"
	"github.com/prometheus/client_golang/prometheus"
	"strings"
	"github.com/prometheus/common/log"
)

// Exporter collects MySQL metrics. It implements prometheus.Collector.
type CollectorCustom struct {
	config          custom_config.MetricsItem
	collectorName   string
	duration, error prometheus.Gauge
	totalScrapes    prometheus.Counter
	scrapeErrors    *prometheus.CounterVec
}

type CollectorGeneric interface {
	prometheus.Collector

	setConfig()
	PromDesc() string
	scrape(ch chan<- prometheus.Metric)
}

func (c *CollectorCustom) setConfig() {
	if len(c.collectorName) < 1 {
		c.collectorName = custom_config.Exporter
	}

	c.duration = prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: custom_config.Namespace,
			Subsystem: c.collectorName,
			Name:      "last_scrape_duration_seconds",
			Help:      "Duration of the last scrape of metrics from " + c.config.Name,
	})

	c.totalScrapes = prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: custom_config.Namespace,
			Subsystem: c.collectorName,
			Name:      "scrapes_total",
			Help:      "Total number of times " + c.config.Name + " was scraped for metrics.",
	})
	c.scrapeErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: custom_config.Namespace,
			Subsystem: c.collectorName,
			Name:      "scrape_errors_total",
			Help:      "Total number of times an error occurred scraping a " + c.config.Name,
	}, []string{"collector"})

	c.error = prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: custom_config.Namespace,
			Subsystem: c.collectorName,
			Name:      "last_scrape_error",
			Help:      "Whether the last scrape of metrics from " + c.config.Name + " resulted in an error (1 for error, 0 for success).",
	})
}

func (e CollectorCustom) PromDesc() string {
	log.Debugln("Call Generic PromDesc")
	return prometheus.BuildFQName(
		custom_config.Namespace,
		e.collectorName,
		strings.ToLower(e.config.Name),
	)
}
