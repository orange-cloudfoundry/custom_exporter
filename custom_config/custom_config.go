package custom_config

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// Metric name parts.
const (
	// Namespace for all metrics.
	Namespace = "custom"
	// Subsystem(s).
	Exporter = "exporter"
)

type CredentialsItem struct {
	Name      string `yaml:"name"`
	Collector string `yaml:"type"`

	Dsn  string `yaml:"dsn,omitempty"`
	Uri  string `yaml:"uri,omitempty"`
	Path string `yaml:"path,omitempty"`

	//@TODO add user to allow run command as this user... for shell need uid/gid
}

type MetricsItem struct {
	Name     string
	Commands []string

	Credential  CredentialsItem

	Mapping    []string
	Separator  string
	value_type string
	Value_type prometheus.ValueType
}

type MetricsItemYaml struct {
	Name     string   `yaml:"name"`
	Commands []string `yaml:"commands"`

	Credential string `yaml:"credential"`

	Mapping    []string `yaml:"mapping"`
	Separator  string   `yaml:"separator,omitempty"`
	Value_type string   `yaml:"value_type"`
}

type ConfigYaml struct {
	Custom_exporter struct {
		Credentials []CredentialsItem `yaml:"credentials"`
		Metrics     []MetricsItemYaml `yaml:"metrics"`
	} `yaml:"custom_exporter"`
}

type Config struct {
	Metrics map[string]MetricsItem
}

func NewConfig(configFile string) (*Config, error) {
	var contentFile []byte
	var err error

	if contentFile, err = ioutil.ReadFile(configFile); err != nil {
		return nil, err
	}

	ymlCnf := ConfigYaml{}

	if err = yaml.Unmarshal(contentFile, &ymlCnf); err != nil {
		return nil, err
	}

	myCnf := new(Config)
	myCnf.metricsList(ymlCnf)

	log.Debugln("config loaded:\n", string(contentFile))

	return myCnf, nil
}

func (c Config) credentialsList(yaml ConfigYaml) map[string]CredentialsItem {
	var result map[string]CredentialsItem

	result = make(map[string]CredentialsItem, 0)

	for _, v := range yaml.Custom_exporter.Credentials {
		result[v.Name] = CredentialsItem{
			Name:      v.Name,
			Collector: v.Collector,
			Dsn:       v.Dsn,
			Path:      v.Path,
			Uri:       v.Uri,
		}
	}

	return result
}

func (e Config) ValueType(Value_type string) prometheus.ValueType {

	switch Value_type {
	case "COUNTER":
		return prometheus.CounterValue
	case "GAUGE":
		return prometheus.GaugeValue
	}

	return prometheus.UntypedValue
}

func (c *Config) metricsList(yaml ConfigYaml) {
	var result map[string]MetricsItem
	var credentials map[string]CredentialsItem

	result = make(map[string]MetricsItem, 0)
	credentials = c.credentialsList(yaml)

	for _, v := range yaml.Custom_exporter.Metrics {
		if cred, ok := credentials[v.Credential]; ok {
			result[v.Name] = MetricsItem{
				Name:        v.Name,
				Commands:    v.Commands,
				Credential:  cred,
				Mapping:     v.Mapping,
				Separator:   v.Separator,
				Value_type:  c.ValueType(v.Value_type),
			}
		} else {
			log.Fatalf("error credential, collector type not found : %s", v.Credential)
		}
	}

	c.Metrics = result
}

func (m MetricsItem) SeparatorValue() string {
	sep := m.Separator

	if len(sep) < 1 {
		sep = "\t"
	}

	return sep
}
