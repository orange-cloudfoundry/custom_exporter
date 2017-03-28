package collector_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"sync"

	"database/sql"
	"database/sql/driver"
	"errors"
	"github.com/orange-cloudfoundry/custom_exporter/collector"
	"github.com/orange-cloudfoundry/custom_exporter/config"
	"github.com/prometheus/common/log"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
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

var _ = Describe("Testing Custom Export, Staging Config Test: ", func() {
	var (
		cnf      *config.Config
		colMysql *collector.CollectorMysql
		metric   config.MetricsItem

		DBclient *sql.DB
		DBmock   sqlmock.Sqlmock

		isOk bool
		err  error
	)

	BeforeEach(func() {
		wg = sync.WaitGroup{}
		wg.Add(1)

		cnf, err = config.NewConfig("../example_with_error.yml")

		if DBclient, DBmock, err = sqlmock.New(); err != nil {
			log.Fatalf("Error while trying to mock DB Mysql connection : %v", err)
		}
	})

	Context("When giving a valid config file with custom_metric_mysql", func() {

		It("should have a valid config object", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		Context("And giving an invalid config metric object", func() {
			It("should found the invalid metric object", func() {
				metric, isOk = cnf.Metrics["custom_metric_shell"]
				Expect(isOk).To(BeTrue())
			})
			It("should return an error when creating the collector", func() {
				_, err = collector.NewPrometheusMysqlCollector(metric)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("And giving an valid config metric object with invalid command", func() {
			It("should found the valid metric object", func() {
				metric, isOk = cnf.Metrics["custom_metric_mysql_error"]
				Expect(isOk).To(BeTrue())
			})

			It("should not return an error when creating the collector", func() {
				_, err = collector.NewPrometheusMysqlCollector(metric)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return a valid Bash collector", func() {
				colMysql = collector.NewCollectorMysql(metric)
				Expect(colMysql.Config()).To(Equal(metric))
				Expect(colMysql.Name()).To(Equal(collector.CollectorMysqlName))
				Expect(colMysql.Desc()).To(Equal(collector.CollectorMysqlDesc))
			})

			It("should return an error when call Run", func() {
				colMysql.StoreDBClient(DBclient)
				DBmock.ExpectQuery("SELECT \"id\", \"name\", 1, FROM animals").WithArgs().WillReturnError(errors.New("Generated SQL Error in mock object"))

				go func() {
					defer func() {
						GinkgoRecover()
						wg.Done()
					}()
					log.Infoln("Calling Run")
					Expect(colMysql.Run(ch)).To(HaveOccurred())
					log.Infoln("Run called...")
				}()
				wg.Wait()
			})
		})

		Context("And giving a valid config metric object", func() {
			It("should found the valid metric object", func() {
				metric, isOk = cnf.Metrics["custom_metric_mysql"]
				Expect(isOk).To(BeTrue())
			})

			It("should not return an error when creating the collector", func() {
				_, err = collector.NewPrometheusMysqlCollector(metric)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return a valid mysql collector", func() {
				colMysql = collector.NewCollectorMysql(metric)
				Expect(colMysql.Config()).To(Equal(metric))
				Expect(colMysql.Name()).To(Equal(collector.CollectorMysqlName))
				Expect(colMysql.Desc()).To(Equal(collector.CollectorMysqlDesc))
			})

			It("should not return an error when call Run", func() {
				colMysql.StoreDBClient(DBclient)

				var rows *sqlmock.Rows
				var rowValues []driver.Value

				rows = sqlmock.NewRows([]string{"id", "name", "count"})

				rowValues = make([]driver.Value, 0)
				rowValues = append(rowValues, 1)
				rowValues = append(rowValues, "chicken")
				rowValues = append(rowValues, 128)
				rows.AddRow(rowValues...)

				rowValues = make([]driver.Value, 0)
				rowValues = append(rowValues, 2)
				rowValues = append(rowValues, "beef")
				rowValues = append(rowValues, 256)
				rows.AddRow(rowValues...)

				rowValues = make([]driver.Value, 0)
				rowValues = append(rowValues, 3)
				rowValues = append(rowValues, "snails")
				rowValues = append(rowValues, 14)
				rows.AddRow(rowValues...)

				DBmock.ExpectQuery("SELECT aml_id,aml_name,aml_number FROM animals").WillReturnRows(rows)

				go func() {
					defer func() {
						GinkgoRecover()
						wg.Done()
					}()
					log.Infoln("Calling Run")
					err := colMysql.Run(ch)

					if err != nil {
						log.Errorf("Error : %v", err)
					}

					Expect(err).ToNot(HaveOccurred())
					log.Infoln("Run called...")
				}()

				wg.Wait()
			})
		})
	})
})
