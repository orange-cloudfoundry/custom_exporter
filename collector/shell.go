package collector

import (
	"fmt"
	"github.com/orange-cloudfoundry/custom_exporter/custom_config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	CollectorShellName = "shell"
	CollectorShellDesc = "Metrics from shell collector in the custom exporter."
)

type CollectorShell struct {
	metricsConfig custom_config.MetricsItem
}

func NewPrometheusShellCollector(config custom_config.MetricsItem) (prometheus.Collector, error) {
	myCol := NewCollectorHelper(&CollectorShell{
		metricsConfig: config,
	})

	log.Infof("Collector Added: Type '%s' / Name '%s' / Credentials '%s'", CollectorShellName, config.Name, config.Credential.Name)

	return myCol, myCol.Check()
}

func (e CollectorShell) Config() custom_config.MetricsItem {
	return e.metricsConfig
}

func (e CollectorShell) Name() string {
	return CollectorShellName
}

func (e CollectorShell) Desc() string {
	return CollectorShellDesc
}

func (e CollectorShell) Run(ch chan <- prometheus.Metric) error {
	var output []byte
	var err error
	var command string

	env := os.Environ()
	env = append(env, fmt.Sprintf("CREDENTIALS_NAME=%s", e.metricsConfig.Credential.Name))
	env = append(env, fmt.Sprintf("CREDENTIALS_COLLECTOR=%s", e.metricsConfig.Credential.Collector))
	env = append(env, fmt.Sprintf("CREDENTIALS_DSN=%s", e.metricsConfig.Credential.Dsn))
	env = append(env, fmt.Sprintf("CREDENTIALS_PATH=%s", e.metricsConfig.Credential.Path))
	env = append(env, fmt.Sprintf("CREDENTIALS_URI=%s", e.metricsConfig.Credential.Uri))

	for _, c := range e.metricsConfig.Commands {

		f := strings.Split(c, " ")

		if len(f) > 1 {
			command, f = f[0], f[1:]
		} else {
			command = f[0]
			f = make([]string, 0)
		}

		_, err = exec.LookPath(command)
		if err != nil {
			log.Errorf("Error whil running command : %s", err.Error())
			return err
		}

		cmd := exec.Command(command, f...)
		cmd.Env = env

		output, err = cmd.Output()

		if err != nil {
			log.Errorf("Error whil running command : %s", err.Error())
			return err
		}
	}

	log.Debugf("Run command '%s', result:", command)
	log.Debugln("Run result:", "\n" + string(output))

	return e.parse(ch, string(output))
}

func (e CollectorShell) parse(ch chan <- prometheus.Metric, output string) error {
	var err error

	err = nil
	sep := e.metricsConfig.Separator
	nb := len(e.metricsConfig.Mapping) + 1

	if len(sep) < 1 {
		sep = "\t"
	}

	for _, l := range strings.Split(output, "\n") {
		if len(strings.TrimSpace(l)) < nb {
			continue
		}

		// prevents first and last char are a separator
		l = strings.Trim(strings.TrimSpace(l), sep)

		if errline := e.parseLine(ch, strings.Split(l, sep)); errline != nil {
			log.Errorf("Error whil parsing line : %s", errline.Error())
			err = errline
		}
	}

	return err
}

func (e *CollectorShell) parseLine(ch chan <- prometheus.Metric, fields []string) error {
	log.Debugln("Call Shell parseLine")
	var (
		mapping   []string
		labelVal  []string
		metricVal float64
		err 	  error
	)

	mapping = e.metricsConfig.Mapping
	labelVal = make([]string, len(mapping))
	err = nil

	for i, value := range fields {

		value = strings.TrimSpace(value)

		if (i + 1) > len(mapping) {
			if metricVal, err = strconv.ParseFloat(value, 64); err != nil {
				metricVal = float64(0)
			}
		} else {
			labelVal[i] = value
		}
	}

	log.Debugf("Add Metric Tag '%s' / TagValue '%s' / Value '%v'", mapping, labelVal, metricVal)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(PromDesc(e), CollectorShellDesc, mapping, nil),
		e.metricsConfig.Value_type, metricVal, labelVal...,
	)

	return err
}
