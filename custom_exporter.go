package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"errors"
	"github.com/orange-cloudfoundry/custom_exporter/collector"
	"github.com/orange-cloudfoundry/custom_exporter/custom_config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
)

var showVersion = flag.Bool(
	"version",
	false,
	"Print version information.",
)

var listenAddress = flag.String(
	"web.listen-address",
	":9209",
	"Address to listen on for web interface and telemetry.",
)

var metricPath = flag.String(
	"web.telemetry-path",
	"/metrics",
	"Path under which to expose metrics.",
)

var configFile = flag.String(
	"collector.config",
	"",
	"Path to config.yml file to read custom exporter definition.",
)

func createListCollectors(c *custom_config.Config) []prometheus.Collector {
	var result []prometheus.Collector

	for _, cnf := range c.Metrics {
		if col := createNewCollector(&cnf); col != nil {
			result = append(result, col)
		}
	}

	if len(result) < 1 {
		log.Fatalf("Error : the metrics list is empty !!")
	}

	return result
}

func createNewCollector(m *custom_config.MetricsItem) prometheus.Collector {
	var col prometheus.Collector
	var err error

	switch m.Credential.Collector {
	case "bash":
		col, err = collector.NewPrometheusBashCollector(*m)
	case "mysql":
		col, err = collector.NewPrometheusMysqlCollector(*m)
	case "redis":
		col, err = collector.NewPrometheusRedisCollector(*m)
	default:
		return nil
	}

	if err != nil {
		log.Errorf("Error:", err.Error())
		return nil
	}

	return col
}

func init() {
	prometheus.MustRegister(version.NewCollector(custom_config.Namespace + "_" + custom_config.Exporter))
}

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Fprintln(os.Stdout, version.Print(custom_config.Namespace+"_"+custom_config.Exporter))
		os.Exit(0)
	}

	if len(*configFile) < 1 {
		err := errors.New("Config file parameter must be provided")
		log.Fatalln("Error:", err.Error())
	}

	if _, err := os.Stat(*configFile); err != nil {
		log.Fatalln("Error:", err.Error())
	}

	log.Infoln("Starting "+custom_config.Namespace+"_"+custom_config.Exporter, version.Info())
	log.Infoln("Build context", version.BuildContext())

	var myConfig *custom_config.Config

	if cnf, err := custom_config.NewConfig(*configFile); err != nil {
		log.Fatalf("FATAL: %s", err.Error())
	} else {
		myConfig = cnf
	}

	prometheus.MustRegister(createListCollectors(myConfig)...)

	http.Handle(*metricPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html><head><title>Custom exporter</title></head><body><h1>Custom exporter</h1><p><a href='` + *metricPath + `'>Metrics</a></p></body></html>`))
	})

	log.Infoln("Listening on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
