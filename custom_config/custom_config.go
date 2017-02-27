package custom_config

import (
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
	Name      string `yaml: "name"`
	Collector string `yaml : "type"`

	Dsn  string `yaml: "dsn, omitempty"`
	Uri  string `yaml: "uri, omitempty"`
	Path string `yaml: "path", omitempty"`

	//@TODO add user to allow run command as this user... for shell need uid/gid
}

type MetricsItem struct {
	Name     string   `yaml: "name"`
	Commands []string `yaml: "commands"`

	credential  string `yaml: "credential"`
	Credentials CredentialsItem

	Mapping    []string `yaml: "mapping"`
	Separator  string   `yaml: "separator, omitempty"`
	Value_type string   `yaml: "value_type"`
}

type configYaml struct {
	credentials []CredentialsItem `yaml: "credentials, flow"`
	metrics     []MetricsItem     `yaml: "metrics, flow"`
}

type Config struct {
	Metrics map[string]MetricsItem
}

func NewConfig(configFile string) *Config {

	var contentFile []byte
	var err error

	if contentFile, err = ioutil.ReadFile(configFile); err != nil {
		log.Fatalf("error while reading file %s : %s", configFile, err.Error())
	}

	ymlCnf := new(configYaml)

	if err = yaml.Unmarshal(contentFile, ymlCnf); err != nil {
		log.Fatalf("error read yaml from file %s : %s", configFile, err.Error())
	}

	myCnf := new(Config)
	myCnf.metricsList(*ymlCnf)

	return myCnf
}

func (c Config) credentialsList(yaml configYaml) map[string]CredentialsItem {
	var result map[string]CredentialsItem

	result = make(map[string]CredentialsItem, 0)

	for _, v := range yaml.credentials {
		result[v.Name] = v
	}

	return result
}

func (c *Config) metricsList(yaml configYaml) {
	var result map[string]MetricsItem
	var credentials map[string]CredentialsItem

	result = make(map[string]MetricsItem, 0)
	credentials = c.credentialsList(yaml)

	for _, v := range yaml.metrics {
		if cred, ok := credentials[v.credential]; ok {
			v.Credentials = cred
			result[v.Name] = v
		} else {
			log.Fatalf("error credential, collector type not found : %s", v.credential)
		}
	}

	c.Metrics = result
}
