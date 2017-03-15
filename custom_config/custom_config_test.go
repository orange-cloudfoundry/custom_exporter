package custom_config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/orange-cloudfoundry/custom_exporter/custom_config"
)

var _ = Describe("Testing Custom Export, Staging Config Test: ", func() {
	var (
		filePath string
		config   *custom_config.Config
		err      error
	)

	JustBeforeEach(func() {
		config, err = custom_config.NewConfig(filePath)
	})

	Context("When miss the file config path", func() {
		BeforeEach(func() {
			filePath = ""
		})

		It("shound occures an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("When give a wrong file path", func() {
		BeforeEach(func() {
			filePath = "/test/me/wrong.yml"
		})

		It("shound occures an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("When give a good config file path and wrong yaml formatted", func() {
		BeforeEach(func() {
			filePath = "../wrongYaml.yml"
		})

		It("shound occures an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("When give a good config file path and well formatted yaml", func() {
		BeforeEach(func() {
			filePath = "../example.yml"
		})

		It("shound not occures an error", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return a config struct of 6 metrics", func() {
			Expect(len(config.Metrics)).To(Equal(3))
		})

		It("should return metrics custom_metric_shell well composed", func() {
			name := "custom_metric_shell"

			Expect(config.Metrics[name].Name).To(Equal(name))
			Expect(len(config.Metrics[name].Commands)).To(Equal(3))
			Expect(config.Metrics[name].Credential.Name).To(Equal("shell_root"))
			Expect(config.Metrics[name].Credential.Collector).To(Equal("bash"))
		})

		It("should return metrics custom_metric_shell well composed", func() {
			name := "custom_metric_mysql"

			Expect(config.Metrics[name].Name).To(Equal(name))
			Expect(len(config.Metrics[name].Commands)).To(Equal(1))
			Expect(config.Metrics[name].Credential.Name).To(Equal("mysql_connector"))
			Expect(config.Metrics[name].Credential.Collector).To(Equal("mysql"))
		})

		It("should return metrics custom_metric_shell well composed", func() {
			name := "custom_metric_redis"

			Expect(config.Metrics[name].Name).To(Equal(name))
			Expect(len(config.Metrics[name].Commands)).To(Equal(1))
			Expect(config.Metrics[name].Credential.Name).To(Equal("redis_connector"))
			Expect(config.Metrics[name].Credential.Collector).To(Equal("redis"))
		})
	})
})
