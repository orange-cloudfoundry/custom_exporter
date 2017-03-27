package collector_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/orange-cloudfoundry/custom_exporter/collector"
	"github.com/orange-cloudfoundry/custom_exporter/custom_config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"net/url"
	"sync"
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
		config   *custom_config.Config
		colRedis *collector.CollectorRedis
		collect  prometheus.Collector
		metric   custom_config.MetricsItem

		isOk bool
		err  error
	)

	BeforeEach(func() {

		wg = sync.WaitGroup{}
		wg.Add(1)

		config, err = custom_config.NewConfig("../example_with_error.yml")

		var dsn *url.URL
		for k, m := range config.Metrics {
			if m.Credential.Collector != collector.CollectorRedisName {
				continue
			}

			if dsn, err = url.Parse(m.Credential.Dsn); err != nil {
				continue
			}

			dsn.Host = redisAddr

			m.Credential.Dsn = dsn.String()
			config.Metrics[k] = m
		}
	})

	Context("When giving a valid config file with custom_metric_redis", func() {

		It("should have a valid config object", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		Context("And giving an invalid config metric object", func() {
			It("should found the invalid metric object", func() {
				metric, isOk = config.Metrics["custom_metric_mysql"]
				Expect(isOk).To(BeTrue())
			})
			It("should return an error when creating the collector", func() {
				collect, err = collector.NewPrometheusRedisCollector(metric)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("And giving an valid config metric object with invalid command", func() {
			It("should found the valid metric object", func() {
				metric, isOk = config.Metrics["custom_metric_redis_error"]
				Expect(isOk).To(BeTrue())
			})

			It("should not return an error when creating the collector", func() {
				collect, err = collector.NewPrometheusRedisCollector(metric)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return a valid Bash collector", func() {
				colRedis = collector.NewCollectorRedis(metric)
				Expect(colRedis.Config()).To(Equal(metric))
				Expect(colRedis.Name()).To(Equal(collector.CollectorRedisName))
				Expect(colRedis.Desc()).To(Equal(collector.CollectorRedisDesc))
			})

			It("should return an error when call Run", func() {
				go func() {
					defer func() {
						GinkgoRecover()
						wg.Done()
					}()
					log.Infoln("Calling Run")
					Expect(colRedis.Run(ch)).To(HaveOccurred())
					log.Infoln("Run called...")
				}()

				wg.Wait()
			})
		})

		Context("And giving a valid config metric object", func() {
			It("should found the valid metric object", func() {
				metric, isOk = config.Metrics["custom_metric_redis"]
				Expect(isOk).To(BeTrue())
			})

			It("should not return an error when creating the collector", func() {
				collect, err = collector.NewPrometheusRedisCollector(metric)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return a valid redis collector", func() {
				colRedis = collector.NewCollectorRedis(metric)
				Expect(colRedis.Config()).To(Equal(metric))
				Expect(colRedis.Name()).To(Equal(collector.CollectorRedisName))
				Expect(colRedis.Desc()).To(Equal(collector.CollectorRedisDesc))
			})

			It("should not return an error when call Run", func() {
				go func() {
					defer func() {
						GinkgoRecover()
						wg.Done()
					}()
					log.Infoln("Calling Run")
					Expect(colRedis.Run(ch)).ToNot(HaveOccurred())
					log.Infoln("Run called...")
				}()

				wg.Wait()
			})
		})
	})
})
