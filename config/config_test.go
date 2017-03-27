package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/orange-cloudfoundry/custom_exporter/config"
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
		filePath string
		cnf      *config.Config
		err      error
	)

	JustBeforeEach(func() {
		cnf, err = config.NewConfig(filePath)
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
			Expect(len(cnf.Metrics)).To(Equal(3))
		})

		It("should return metrics custom_metric_shell well composed", func() {
			name := "custom_metric_shell"

			Expect(cnf.Metrics[name].Name).To(Equal(name))
			Expect(len(cnf.Metrics[name].Commands)).To(Equal(3))
			Expect(cnf.Metrics[name].Credential.Name).To(Equal("shell_root"))
			Expect(cnf.Metrics[name].Credential.Collector).To(Equal("bash"))
		})

		It("should return metrics custom_metric_shell well composed", func() {
			name := "custom_metric_mysql"

			Expect(cnf.Metrics[name].Name).To(Equal(name))
			Expect(len(cnf.Metrics[name].Commands)).To(Equal(1))
			Expect(cnf.Metrics[name].Credential.Name).To(Equal("mysql_connector"))
			Expect(cnf.Metrics[name].Credential.Collector).To(Equal("mysql"))
		})

		It("should return metrics custom_metric_shell well composed", func() {
			name := "custom_metric_redis"

			Expect(cnf.Metrics[name].Name).To(Equal(name))
			Expect(len(cnf.Metrics[name].Commands)).To(Equal(1))
			Expect(cnf.Metrics[name].Credential.Name).To(Equal("redis_connector"))
			Expect(cnf.Metrics[name].Credential.Collector).To(Equal("redis"))
		})
	})
})
