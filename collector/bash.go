package collector

import (
	"bytes"
	"fmt"
	"github.com/orange-cloudfoundry/custom_exporter/custom_config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	CollectorBashName = "bash"
	CollectorBashDesc = "Metrics from shell collector in the custom exporter."
)

type CollectorBash struct {
	metricsConfig custom_config.MetricsItem
}

func NewCollectorBash(config custom_config.MetricsItem) *CollectorBash {
	return &CollectorBash{
		metricsConfig: config,
	}
}

func NewPrometheusBashCollector(config custom_config.MetricsItem) (prometheus.Collector, error) {
	myCol := NewCollectorHelper(
		NewCollectorBash(config),
	)

	log.Infof("Collector Added: Type '%s' / Name '%s' / Credentials '%s'", CollectorBashName, config.Name, config.Credential.Name)

	return myCol, myCol.Check(nil)
}

func (e CollectorBash) Config() custom_config.MetricsItem {
	return e.metricsConfig
}

func (e CollectorBash) Name() string {
	return CollectorBashName
}

func (e CollectorBash) Desc() string {
	return CollectorBashDesc
}

func (e CollectorBash) Run(ch chan<- prometheus.Metric) error {
	var output []byte
	var err error
	var command string

	var bufInput bytes.Buffer
	var bufOutput bytes.Buffer

	env := os.Environ()
	env = append(env, fmt.Sprintf("CREDENTIALS_NAME=%s", e.metricsConfig.Credential.Name))
	env = append(env, fmt.Sprintf("CREDENTIALS_COLLECTOR=%s", e.metricsConfig.Credential.Collector))
	env = append(env, fmt.Sprintf("CREDENTIALS_DSN=%s", e.metricsConfig.Credential.Dsn))
	env = append(env, fmt.Sprintf("CREDENTIALS_PATH=%s", e.metricsConfig.Credential.Path))
	env = append(env, fmt.Sprintf("CREDENTIALS_URI=%s", e.metricsConfig.Credential.Uri))

	bufInput.Reset()
	bufOutput.Reset()

	for _, c := range e.metricsConfig.Commands {

		f := strings.Split(c, " ")

		if len(f) > 1 {
			command, f = f[0], f[1:]
		} else {
			command = f[0]
			f = make([]string, 0)
		}

		log.Debugf("Checking command/script exists : \"%s\"...", command)

		_, err = exec.LookPath(command)
		if err != nil {
			log.Errorf("Error with metric \"%s\" while checking command exists \"%s\" : %s", e.metricsConfig.Name, c, err.Error())
			return err
		}

		log.Debugf("Running command \"%s\" with params \"%s\"...", command, strings.Join(f, " "))

		bufInput.Reset()
		io.Copy(&bufInput, &bufOutput)
		bufOutput.Reset()

		cmd := exec.Command(command, f...)
		cmd.Env = env
		cmd.Stdin = &bufInput
		cmd.Stdout = &bufOutput

		err = cmd.Run()

		if err != nil {
			log.Errorf("Error with metric \"%s\" while running command \"%s\" : %s", e.metricsConfig.Name, c, err.Error())
			return err
		}

		output = bufOutput.Bytes()

		log.Debugf("Result command \"%s\" : \"%s\"", command, string(output))
	}

	log.Debugf("Run metric \"%s\" command '%s', result:", e.metricsConfig.Name, command)
	log.Debugln("Result:", "\n"+string(output))

	return e.parse(ch, string(output))
}

func (e CollectorBash) parse(ch chan<- prometheus.Metric, output string) error {
	var err error

	err = nil
	sep := e.metricsConfig.Separator
	nb := len(e.metricsConfig.Mapping) + 1

	for _, l := range strings.Split(output, "\n") {
		if len(strings.TrimSpace(l)) < nb {
			continue
		}

		log.Debugf("Parsing line: \"%s\"...", l)

		// prevents first and last char are a separator
		l = strings.Trim(strings.TrimSpace(l), sep)

		if errline := e.parseLine(ch, strings.Split(l, sep)); errline != nil {
			log.Errorf("Error with metric \"%s\" while parsing line : %s", e.metricsConfig.Name, errline.Error())
			err = errline
		}
	}

	return err
}

func (e *CollectorBash) parseLine(ch chan<- prometheus.Metric, fields []string) error {
	var (
		mapping   []string
		labelVal  []string
		metricVal float64
		err       error
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

	if err != nil {
		log.Debugf("Return error : '%s'", err.Error())
		return err
	}

	prom_desc := PromDesc(e)
	log.Debugf("Add Metric \"%s\" : Tag '%s' / TagValue '%s' / Value '%v'", prom_desc, mapping, labelVal, metricVal)

	metric := prometheus.MustNewConstMetric(
		prometheus.NewDesc(prom_desc, CollectorBashDesc, mapping, nil),
		e.metricsConfig.Value_type, metricVal, labelVal...,
	)

	select {
	case ch <- metric:
		log.Debug("Return no error...")
		return nil
	default:
		log.Info("Cannot write to channel...")
	}

	return err
}
