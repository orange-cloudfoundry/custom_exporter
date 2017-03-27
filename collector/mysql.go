package collector

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/orange-cloudfoundry/custom_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"

	_ "github.com/go-sql-driver/mysql"
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

const (
	CollectorMysqlName = "mysql"
	CollectorMysqlDesc = "Metrics from mysql collector in the custom exporter."
)

type CollectorMysql struct {
	client        *sql.DB
	metricsConfig config.MetricsItem
}

func NewCollectorMysql(config config.MetricsItem) *CollectorMysql {
	return &CollectorMysql{
		metricsConfig: config,
	}
}

func NewPrometheusMysqlCollector(config config.MetricsItem) (prometheus.Collector, error) {
	myCol := NewCollectorHelper(NewCollectorMysql(config))

	log.Infof("Collector Added: Type '%s' / Name '%s' / Credentials '%s'", CollectorMysqlName, config.Name, config.Credential.Name)
	return myCol, myCol.Check(nil)
}

func (e CollectorMysql) Config() config.MetricsItem {
	return e.metricsConfig
}

func (e CollectorMysql) Name() string {
	return CollectorMysqlName
}

func (e CollectorMysql) Desc() string {
	return CollectorMysqlDesc
}

func (e *CollectorMysql) Run(ch chan<- prometheus.Metric) error {
	var (
		err error
		out *sql.Rows
	)

	err = nil

	if e.client == nil {
		if err = e.DBClient(); err != nil {
			log.Errorf("Error for metrics \"%s\" while creating DB client \"%s\": %v", e.metricsConfig.Name, e.metricsConfig.Credential.Dsn, err)
			return err
		}
	}

	defer e.client.Close()

	if err = e.client.Ping(); err != nil {
		log.Errorf("Error for metrics \"%s\" while trying to ping DB server \"%s\": %v", e.metricsConfig.Name, e.metricsConfig.Credential.Dsn, err)
		return err
	}

	log.Debugln("Calling Mysql Commands... ")

	for _, c := range e.metricsConfig.Commands {
		c = strings.TrimSpace(c)

		if len(c) < 1 {
			continue
		}

		if out, err = e.client.Query(c); err != nil {
			log.Errorf("Error for metrics \"%s\" while calling query \"%s\": %v", e.metricsConfig.Name, c, err)
			return err
		}
	}

	return e.parseResult(ch, out)
}

func (e *CollectorMysql) parseResult(ch chan<- prometheus.Metric, res *sql.Rows) error {
	var (
		err        error
		nbCols     int
		colMapping map[int]string
		tagLabels  []string
		tagValues  []string
		valMetric  float64
	)

	if colList, err := res.Columns(); err != nil {
		log.Errorf("Error for metrics \"%s\" while retrieve columns names \"%s\": %v", e.metricsConfig.Name, err)
		return err
	} else {
		nbCols = len(colList)
		colMapping = e.mapColumsConfig(colList, e.metricsConfig.Mapping)
	}

	log.Debugf("Metrics \"%s\" - Colums lists : %v", e.metricsConfig.Name, colMapping)

	for res.Next() {
		ptrMapping := make([]interface{}, nbCols)
		rawMapping := make([][]byte, nbCols)

		tagLabels = make([]string, 0)
		tagValues = make([]string, 0)
		valMetric = float64(0)

		for i, _ := range colMapping {
			if (i + 1) < len(colMapping) {
				ptrMapping[i] = &rawMapping[i]
			} else {
				ptrMapping[i] = &valMetric
			}
		}

		if errRow := res.Scan(ptrMapping...); errRow != nil {
			log.Errorf("Error for metrics \"%s\", while parsing result : %v", e.metricsConfig.Name, errRow)
			err = errRow
			continue
		}

		for i, k := range colMapping {
			if (i + 1) < len(colMapping) {
				if k != "" {
					tagLabels = append(tagLabels, k)
					tagValues = append(tagValues, string(rawMapping[i]))
				}
			}
		}

		prom_desc := PromDesc(e)
		log.Debugf("Add Metric \"%s\" : Tag '%s' / TagValue '%s' / Value '%v'", prom_desc, tagLabels, tagValues, valMetric)

		metric := prometheus.MustNewConstMetric(
			prometheus.NewDesc(prom_desc, e.metricsConfig.Name, tagLabels, nil),
			e.metricsConfig.Value_type, valMetric, tagValues...,
		)

		select {
		case ch <- metric:
			log.Debug("Return no error...")
		default:
			log.Info("Cannot write to channel...")
		}
	}

	return err
}

func (e *CollectorMysql) mapColumsConfig(colums, config []string) map[int]string {
	var res map[int]string

	res = make(map[int]string, 0)

	for i, c := range colums {

		c = strings.TrimSpace(c)

		for _, k := range config {

			k = strings.TrimSpace(k)

			if k == c {
				res[i] = k
			}
		}

		if _, ok := res[i]; !ok {
			res[i] = ""
		}
	}

	return res
}

func (e CollectorMysql) DsnPart() (string, string, error) {
	dsn := strings.TrimSpace(e.metricsConfig.Credential.Dsn)

	if len(dsn) < 1 {
		return "", "", errors.New(fmt.Sprintf("Cannot find a valid dsn : %s", e.metricsConfig.Credential.Dsn))
	}

	dsnPart := strings.SplitN(e.metricsConfig.Credential.Dsn, "://", 2)

	if dsnPart[0] == "" {
		return "", "", errors.New(fmt.Sprintf("Cannot find a valid dsn : %s", e.metricsConfig.Credential.Dsn))
	}

	if len(dsnPart[1]) < 3 {
		return "", "", errors.New(fmt.Sprintf("Cannot find a valid dsn : %s", e.metricsConfig.Credential.Dsn))
	}

	return dsnPart[0], dsnPart[1], nil
}

func (e *CollectorMysql) DBClient() error {
	var (
		dsnstr string
		driver string
		client *sql.DB
		err    error
	)

	if driver, dsnstr, err = e.DsnPart(); err != nil {
		return err
	}

	if client, err = sql.Open(driver, dsnstr); err != nil {
		return err
	}

	e.StoreDBClient(client)

	return nil
}

func (e *CollectorMysql) StoreDBClient(client *sql.DB) {
	e.client = client
}
