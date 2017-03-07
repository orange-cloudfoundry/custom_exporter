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
)

const (
	CollectorShellName = "shell"
	CollectorShellDesc = "Metrics from shell collector in the custom exporter."
)

type CollectorShell struct {
	metricsConfig custom_config.MetricsItem
}

func NewPrometheusShellCollector(config custom_config.MetricsItem) (prometheus.Collector, error) {
	if config.Credential.Collector != CollectorShellName {
		err := errors.New(
			fmt.Sprintf("Error mismatching collector type : config type = %s & current type = %s",
				config.Credential.Collector,
				CollectorShellName,
			))
		log.Fatalln("Error:", err)
		return nil, err
	}

	myCol := NewCollectorHelper(&CollectorShell{
		metricsConfig: config,
	})

	if len(config.Commands) < 1 {
		err := errors.New("Error empty commands to run !!")
		log.Errorln("Error:", err)
		return myCol, err
	}

	log.Infof("Collector Added: Type '%s' / Name '%s' / Credentials '%s'", CollectorShellName, config.Name, config.Credential.Name)
	return myCol, nil
}
func (e CollectorShell) Config() custom_config.MetricsItem {
	return e.metricsConfig
}
func (e CollectorShell) Name() string {
	return CollectorShellName
}
func (e CollectorShell) Run(ch chan <- prometheus.Metric) error {
	log.Debugln("Call Shell run")

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
			log.Fatalf("Error whil running command : %s", err.Error())
			return err
		}

		cmd := exec.Command(command, f...)
		cmd.Env = env

		output, err = cmd.Output()

		if err != nil {
			log.Fatalf("Error whil running command : %s", err.Error())
			return err
		}
	}

	log.Debugf("Run command '%s', result:", command)
	log.Debugln("Run result:", "\n" + string(output))

	return e.parse(ch, string(output))
}

func (e CollectorShell) parse(ch chan <- prometheus.Metric, output string) error {
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

		e.parseLine(ch, strings.Split(l, sep))
	}

	return nil
}

func (e *CollectorShell) parseLine(ch chan <- prometheus.Metric, fields []string) {
	log.Debugln("Call Shell parseLine")
	var (
		mapping   []string
		labelVal  []string
		metricVal float64
	)

	mapping = e.metricsConfig.Mapping
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

	log.Debugf("Add Metric Tag '%s' / TagValue '%s' / Value '%v'", mapping, labelVal, metricVal)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(PromDesc(e), CollectorShellDesc, mapping, nil),
		e.metricsConfig.Value_type, metricVal, labelVal...,
	)
}
