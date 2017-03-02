package collector

import (
	"errors"
	"fmt"
	"github.com/orange-cloudfoundry/custom_exporter/custom_config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	CollectorShellName = "shell"
	CollectorShellDesc = "Metrics from shell collector in the custom exporter."
)

type CollectorShell struct {
	CollectorCustom
}

func NewShell(config custom_config.MetricsItem) (*CollectorShell, error) {
	myCol := CollectorShell{CollectorCustom{config:config,collectorName:CollectorShellName}}
	myCol.setConfig()

	if myCol.config.Credential.Collector != CollectorShellName {
		err := errors.New(
			fmt.Sprintf("Error mismatching collector type : config type = %s & current type = %s",
				myCol.config.Credential.Collector,
				CollectorShellName,
			))
		log.Fatalln("Error:", err)
		return nil, err
	}

	if len(myCol.config.Commands) < 1 {
		err := errors.New("Error empty commands to run !!")
		log.Errorln("Error:", err)
		return &myCol, err
	}

	log.Infof("Collector Added: Type '%s' / Name '%s' / Credentials '%s'", myCol.CollectorCustom.collectorName, myCol.CollectorCustom.config.Name, myCol.CollectorCustom.config.Credential.Name)
	return &myCol, nil
}

func (e *CollectorShell) Describe(ch chan<- *prometheus.Desc) {
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
func (e *CollectorShell) Collect(ch chan<- prometheus.Metric) {
	log.Debugln("Call Generic Collect")
	e.scrape(ch)
	ch <- e.duration
	ch <- e.totalScrapes
	ch <- e.error
	e.scrapeErrors.Collect(ch)
}

func (e *CollectorShell) scrape(ch chan<- prometheus.Metric) {
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

	if out, err := e.run(); err == nil {
		e.parse(ch, out)
	}
}

func (e CollectorShell) run() (string, error) {
	log.Debugln("Call Shell run")

	var output []byte
	var err error
	var command string

	env := os.Environ()
	env = append(env, fmt.Sprintf("CREDENTIALS_NAME=%s", e.config.Credential.Name))
	env = append(env, fmt.Sprintf("CREDENTIALS_COLLECTOR=%s", e.config.Credential.Collector))
	env = append(env, fmt.Sprintf("CREDENTIALS_DSN=%s", e.config.Credential.Dsn))
	env = append(env, fmt.Sprintf("CREDENTIALS_PATH=%s", e.config.Credential.Path))
	env = append(env, fmt.Sprintf("CREDENTIALS_URI=%s", e.config.Credential.Uri))

	for _, c := range e.config.Commands {

		f := strings.Split(c, " ")

		if len(f) > 1 {
			command, f = f[0], f[1:]
		} else {
			command = f[0]
			f = make([]string, 0)
		}

		_, err = exec.LookPath(command)
		if err != nil {
			log.Fatalf("Error whil running command : %s", err.Error())
			return "", err
		}

		cmd := exec.Command(command, f...)
		cmd.Env = env

		output, err = cmd.Output()

		if err != nil {
			log.Fatalf("Error whil running command : %s", err.Error())
			return "", err
		}
	}

	log.Debugf("Run command '%s', result:", command)
	log.Debugln("Run result:", "\n" + string(output))

	return string(output), nil
}

func (e CollectorShell) parse(ch chan<- prometheus.Metric, output string) {
	log.Debugln("Call Shell parse")
	sep := e.config.Separator

	nb := len(e.config.Mapping) + 1

	if len(sep) < 1 {
		sep = "\t"
	}

	for _, l := range strings.Split(output, "\n") {
		if len(strings.TrimSpace(l)) < nb {
			continue
		}

		// prevents first and last char are a separator
		l = strings.Trim(strings.TrimSpace(l), sep)

		e.parseLine(ch, strings.Split(l, sep))
	}
}

func (e CollectorShell) parseLine(ch chan<- prometheus.Metric, fields []string) {
	log.Debugln("Call Shell parseLine")
	var (
		mapping   []string
		labelVal  []string
		metricVal float64
	)

	mapping = e.config.Mapping
	labelVal = make([]string, len(mapping))

	for i, value := range fields {

		value = strings.TrimSpace(value)

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

	log.Debugf("Add Metric Tag '%s' / TagValue '%s' / Value '%s'", mapping, labelVal, metricVal)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(e.PromDesc(), CollectorShellDesc, mapping, nil),
		e.config.Value_type, metricVal, labelVal...,
	)
}
