package collector_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/orange-cloudfoundry/custom_exporter/collector"
	"github.com/orange-cloudfoundry/custom_exporter/custom_config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

var _ = Describe("Testing Custom Export, Staging Config Test: ", func() {
	var (
		config  *custom_config.Config
		colBash *collector.CollectorBash
		collect prometheus.Collector
		metric  custom_config.MetricsItem

		isOk bool
		err  error
	)

	BeforeEach(func() {
		config, err = custom_config.NewConfig("../example_with_error.yml")
	})

	Context("When giving a valid config file with custom_metric_shell", func() {

		It("should have a valid config object", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		Context("And giving an invalid config metric object", func() {
			It("should found the invalid metric object", func() {
				metric, isOk = config.Metrics["custom_metric_mysql"]
				Expect(isOk).To(BeTrue())
			})
			It("should return an error when creating the collector", func() {
				collect, err = collector.NewPrometheusBashCollector(metric)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("And giving an valid config metric object with invalid command", func() {
			It("should found the valid metric object", func() {
				metric, isOk = config.Metrics["custom_metric_shell_error"]
				Expect(isOk).To(BeTrue())
			})

			It("should not return an error when creating the collector", func() {
				collect, err = collector.NewPrometheusBashCollector(metric)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return a valid Bash collector", func() {
				colBash = collector.NewCollectorBash(metric)
				Expect(colBash.Config()).To(Equal(metric))
				Expect(colBash.Name()).To(Equal(collector.CollectorBashName))
				Expect(colBash.Desc()).To(Equal(collector.CollectorBashDesc))
			})

			It("should return an error when call Run", func() {
				log.Debugln("Calling Run")
				Expect(colBash.Run(ch)).To(HaveOccurred())
				log.Debugln("Run called...")
			})

			It("should have capabilities to call the Collect funct", func() {
				go func() {
					defer GinkgoRecover()
					log.Debugln("Calling Collect")
					collect.Collect(ch)
					log.Debugln("Collect called...")
				}()
				<-ch
			})

			It("should have capabilities to call the Describe funct", func() {
				go func() {
					defer GinkgoRecover()
					log.Debugln("Calling Describe")
					collect.Describe(ds)
					log.Debugln("Describe called...")
				}()
				<-ch
			})
		})

		Context("And giving a valid config metric object", func() {
			It("should found the valid metric object", func() {
				metric, isOk = config.Metrics["custom_metric_shell"]
				Expect(isOk).To(BeTrue())
			})

			It("should not return an error when creating the collector", func() {
				collect, err = collector.NewPrometheusBashCollector(metric)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return a valid Bash collector", func() {
				colBash = collector.NewCollectorBash(metric)
				Expect(colBash.Config()).To(Equal(metric))
				Expect(colBash.Name()).To(Equal(collector.CollectorBashName))
				Expect(colBash.Desc()).To(Equal(collector.CollectorBashDesc))
			})

			It("should not return an error when call Run", func() {
				go func() {
					defer GinkgoRecover()
					log.Debugln("Calling Run")
					Expect(colBash.Run(ch)).ToNot(HaveOccurred())
					log.Debugln("Run called...")
				}()
				<-ch
			})

			It("should have capabilities to call the Collect funct", func() {
				go func() {
					defer GinkgoRecover()
					log.Debugln("Calling Collect")
					collect.Collect(ch)
					log.Debugln("Collect called...")
				}()
				<-ch
			})
			It("should have capabilities to call the Describe funct", func() {
				go func() {
					defer GinkgoRecover()
					log.Debugln("Calling Describe")
					collect.Describe(ds)
					log.Debugln("Describe called...")
				}()
				<-ch
			})
		})
	})
})
