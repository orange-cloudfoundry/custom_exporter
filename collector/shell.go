package collector

import (
	"fmt"
	"github.com/orange-cloudfoundry/custom_exporter/custom_config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"os"
	"os/exec"
	"strings"
	"time"
	"errors"
	"strconv"
)

const (
	collectorShellName     = "shell"
	collectorShellDesc = "Metrics from shell collector in the custom exporter."
)

type collectorShell struct {
	*CollectorCustom
}

func NewShell(config *custom_config.MetricsItem) (*collectorShell, error) {
	var myCol *collectorShell
	var err error

	if myCol.CollectorCustom, err = NewCollectorCustom(config, collectorShellName); err != nil {
		return nil, err
	}

	if myCol.config.Credentials.Collector != collectorShellName {
		err := errors.New(
			fmt.Sprintf("Error mismatching collector type : config type = %s & current type = %s",
				myCol.config.Credentials.Collector,
				collectorShellName,
			))
		return nil, err
	}

	if len(myCol.config.Commands) < 1 {
		err := errors.New("Error empty commands to run !!")
		return nil, err
	}

	return myCol, nil
}

func (e *collectorShell) scrape(ch chan<- prometheus.Metric) {

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

	if out, err := e.run(); err != nil {
		log.Fatalf("Error whil running command : %s", err.Error())
		return
	} else {
		e.parse(ch, out)
	}

}

func (e collectorShell) run() (string, error) {

	var output []byte
	var err error

	env := os.Environ()
	env = append(env, fmt.Sprintf("CREDENTIALS_NAME=%s", e.config.Credentials.Name))
	env = append(env, fmt.Sprintf("CREDENTIALS_COLLECTOR=%s", e.config.Credentials.Collector))
	env = append(env, fmt.Sprintf("CREDENTIALS_DSN=%s", e.config.Credentials.Dsn))
	env = append(env, fmt.Sprintf("CREDENTIALS_PATH=%s", e.config.Credentials.Path))
	env = append(env, fmt.Sprintf("CREDENTIALS_URI=%s", e.config.Credentials.Uri))

	for _, c := range e.config.Commands {

		_, err = exec.LookPath(c)
		if err != nil {
			return "", err
		}

		cmd := exec.Command(c)
		cmd.Env = env

		output, err = cmd.Output()

		if err != nil {
			return "", err
		}
	}

	return fmt.Sprint(output), nil
}

func (e collectorShell) parse(ch chan<- prometheus.Metric, output string) {
	sep := e.config.Separator
	nb := len(e.config.Mapping) + 1

	if len(sep) < 1 {
		sep = "\t"
	}

	for _, l := range strings.Split(output, "\n") {
		e.parseLine(ch, strings.SplitN(l, sep, nb))
	}
}

func (e collectorShell) parseLine(ch chan<- prometheus.Metric, lines []string) {
	var (
		mapping   []string
		labelVal  []string
		metricVal float64
	)

	mapping = e.config.Mapping
	labelVal = make([]string, len(mapping))

	for i, value := range lines {
		if (i + 1) > len(mapping) {

			if val, err := strconv.ParseFloat(value, 64); err == nil {
				metricVal = val
			} else {
				metricVal = float64(0)
			}

		} else {
			labelVal[i] = value
		}
	}

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(e.PromDesc(), collectorShellDesc, mapping, nil),
		e.ValueType(), metricVal, labelVal...,
	)
}
