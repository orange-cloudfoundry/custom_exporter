package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/orange-cloudfoundry/custom_exporter/collector"
	"github.com/orange-cloudfoundry/custom_exporter/custom_config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
)

var (
	showVersion = flag.Bool(
		"version",
		false,
		"Print version information.",
	)
	listenAddress = flag.String(
		"web.listen-address",
		":9209",
		"Address to listen on for web interface and telemetry.",
	)
	metricPath = flag.String(
		"web.telemetry-path",
		"/metrics",
		"Path under which to expose metrics.",
	)
	configFile = flag.String(
		"config.yml",
		path.Join(os.Getenv("HOME"), "config.yml"),
		"Path to config.yml file to read custom exporter definition.",
	)
)

// landingPage contains the HTML served at '/'.
// TODO: Make this nicer and more informative.
var landingPage = []byte(`<html>
<head><title>Custom exporter</title></head>
<body>
<h1>Custom exporter</h1>
<p><a href='` + *metricPath + `'>Metrics</a></p>
</body>
</html>
`)

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

	switch m.Credentials.Collector {
	case "shell":
		col, err = collector.NewShell(m)
	}

	if err != nil {
		log.Errorf("Error : %s", err.Error())
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

	log.Infoln("Starting "+custom_config.Namespace+"_"+custom_config.Exporter, version.Info())
	log.Infoln("Build context", version.BuildContext())

	myConfig := custom_config.NewConfig(*configFile)

	prometheus.MustRegister(createListCollectors(myConfig)...)

	http.Handle(*metricPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(landingPage)
	})

	log.Infoln("Listening on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
