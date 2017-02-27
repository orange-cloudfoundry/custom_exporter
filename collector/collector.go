package collector

import (
	"github.com/orange-cloudfoundry/custom_exporter/custom_config"
	"github.com/prometheus/client_golang/prometheus"
	"strings"
	"time"
)

// Exporter collects MySQL metrics. It implements prometheus.Collector.
type CollectorCustom struct {
	config          custom_config.MetricsItem
	collectorName   string
	duration, error prometheus.Gauge
	totalScrapes    prometheus.Counter
	scrapeErrors    *prometheus.CounterVec
}

func NewCollectorCustom(config *custom_config.MetricsItem, collectorName string) (*CollectorCustom, error) {

	if len(collectorName) < 1 {
		collectorName = custom_config.Exporter
	}

	return &CollectorCustom{
		config:        *config,
		collectorName: collectorName,
		duration: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: custom_config.Namespace,
			Subsystem: collectorName,
			Name:      "last_scrape_duration_seconds",
			Help:      "Duration of the last scrape of metrics from " + config.Name,
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: custom_config.Namespace,
			Subsystem: collectorName,
			Name:      "scrapes_total",
			Help:      "Total number of times " + config.Name + " was scraped for metrics.",
		}),
		scrapeErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: custom_config.Namespace,
			Subsystem: collectorName,
			Name:      "scrape_errors_total",
			Help:      "Total number of times an error occurred scraping a " + config.Name,
		}, []string{"collector"}),
		error: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: custom_config.Namespace,
			Subsystem: collectorName,
			Name:      "last_scrape_error",
			Help:      "Whether the last scrape of metrics from " + config.Name + " resulted in an error (1 for error, 0 for success).",
		}),
	}, nil
}

func (e CollectorCustom) ValueType() prometheus.ValueType {

	switch e.config.Value_type {
	case "UNTYPED":
		return prometheus.UntypedValue
	case "COUNTER":
		return prometheus.CounterValue
	case "GAUGE":
		return prometheus.GaugeValue
	}

	return prometheus.UntypedValue
}

func (e CollectorCustom) PromDesc() string {
	return prometheus.BuildFQName(
		custom_config.Namespace,
		e.collectorName,
		strings.ToLower(e.config.Name),
	)
}

// Describe implements prometheus.Collector.
func (e *CollectorCustom) Describe(ch chan<- *prometheus.Desc) {
	// We cannot know in advance what metrics the exporter will generate
	// from MySQL. So we use the poor man's describe method: Run a collect
	// and send the descriptors of all the collected metrics. The problem
	// here is that we need to connect to the MySQL DB. If it is currently
	// unavailable, the descriptors will be incomplete. Since this is a
	// stand-alone exporter and not used as a library within other code
	// implementing additional metrics, the worst that can happen is that we
	// don't detect inconsistent metrics created by this exporter
	// itself. Also, a change in the monitored MySQL instance may change the
	// exported metrics during the runtime of the exporter.

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
func (e *CollectorCustom) Collect(ch chan<- prometheus.Metric) {
	e.scrape(ch)

	ch <- e.duration
	ch <- e.totalScrapes
	ch <- e.error
	e.scrapeErrors.Collect(ch)
}

func (e *CollectorCustom) scrape(ch chan<- prometheus.Metric) {

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
}
