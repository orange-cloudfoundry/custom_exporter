package config

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os/user"
	"strconv"
	"strings"
)

/*
Copyright 2017 Orange

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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

	User string `yaml:"user,omitempty"`
	Dsn  string `yaml:"dsn,omitempty"`
	Uri  string `yaml:"uri,omitempty"`
	Path string `yaml:"path,omitempty"`

	//@TODO add user to allow run command as this user... for shell need uid/gid
}

type MetricsItem struct {
	Name     string
	Commands []string

	Credential CredentialsItem

	Mapping    []string
	Separator  string
	Value_name string
	Value_type prometheus.ValueType
}

type MetricsItemYaml struct {
	Name     string   `yaml:"name"`
	Commands []string `yaml:"commands"`

	Credential string `yaml:"credential"`

	Mapping    []string `yaml:"mapping"`
	Separator  string   `yaml:"separator,omitempty"`
	Value_name string   `yaml:"value_name,omitempty"`
	Value_type string   `yaml:"value_type"`
}

type ConfigYaml struct {
	Credentials []CredentialsItem `yaml:"credentials"`
	Metrics     []MetricsItemYaml `yaml:"metrics"`
}

type Config struct {
	Metrics map[string]MetricsItem
}

type CredentialsUser struct {
	user.User
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

	//log.Debugln("config loaded:\n", string(contentFile))

	return myCnf, nil
}

func (c Config) credentialsList(yaml ConfigYaml) map[string]CredentialsItem {
	var result map[string]CredentialsItem

	result = make(map[string]CredentialsItem, 0)

	for _, v := range yaml.Credentials {
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

	for _, v := range yaml.Metrics {
		if cred, ok := credentials[v.Credential]; ok {
			result[v.Name] = MetricsItem{
				Name:       v.Name,
				Commands:   v.Commands,
				Credential: cred,
				Mapping:    v.Mapping,
				Separator:  v.Separator,
				Value_name: v.Value_name,
				Value_type: c.ValueType(v.Value_type),
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
		sep = " "
	}

	return sep
}

func (m MetricsItem) CredentialUser() *CredentialsUser {
	usr := strings.TrimSpace(m.Credential.User)

	if len(usr) == 0 {
		return currentUser()
	}

	if myUser, err := user.LookupId(usr); err == nil {
		return &CredentialsUser{User: *myUser}
	}

	if myUser, err := user.Lookup(usr); err == nil {
		return &CredentialsUser{User: *myUser}
	}

	return currentUser()
}

func currentUser() *CredentialsUser {
	var myUser *user.User
	var err error

	if myUser, err = user.Current(); err != nil {
		log.Fatalf("Error on retrieve current system user : %s", err.Error())
	}

	return &CredentialsUser{User: *myUser}
}

func (c CredentialsUser) UidInt() uint32 {
	if uid, err := strconv.ParseUint(c.Uid, 10, 32); err == nil {
		return uint32(uid)
	}
	return 0
}

func (c CredentialsUser) GidInt() uint32 {
	if gid, err := strconv.ParseUint(c.Gid, 10, 32); err == nil {
		return uint32(gid)
	}
	return 0
}
