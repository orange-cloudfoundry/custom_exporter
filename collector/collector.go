package collector

import (
	"github.com/orange-cloudfoundry/custom_exporter/custom_config"
	"github.com/prometheus/client_golang/prometheus"
)

const CollectorType = "Collector_Type"
const CollectorConfig = "Collector_Config"

// Exporter collects MySQL metrics. It implements prometheus.Collector.
type CollectorCustom struct {
	config          custom_config.MetricsItem
	duration, error prometheus.Gauge
	totalScrapes    prometheus.Counter
	scrapeErrors    *prometheus.CounterVec
}

func NewCollectorCustom(config custom_config.MetricsItem) (*CollectorCustom, error) {
	return  &CollectorCustom{
		config: config,
		duration: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: custom_config.Namespace,
			Subsystem: custom_config.Exporter,
			Name:      "last_scrape_duration_seconds",
			Help:      "Duration of the last scrape of metrics from "+ config.Name,
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: custom_config.Namespace,
			Subsystem: custom_config.Exporter,
			Name:      "scrapes_total",
			Help:      "Total number of times "+ config.Name +" was scraped for metrics.",
		}),
		scrapeErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: custom_config.Namespace,
			Subsystem: custom_config.Exporter,
			Name:      "scrape_errors_total",
			Help:      "Total number of times an error occurred scraping a "+ config.Name,
		}, []string{"collector"}),
		error: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: custom_config.Namespace,
			Subsystem: custom_config.Exporter,
			Name:      "last_scrape_error",
			Help:      "Whether the last scrape of metrics from "+ config.Name +" resulted in an error (1 for error, 0 for success).",
		}),
	}, nil
}