package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"os"
	"strings"
	"testing"
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

var binaryPath string

func init() {
	find := false

	for _, v := range os.Args {
		if strings.Contains(v, "log.level") {
			find = true
			break
		}
	}

	if !find {
		os.Args = append(os.Args, "-log.level=debug")
	}
}

func TestCustomExporter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Custom Exporter Main Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	var err error
	binaryPath, err = gexec.Build("github.com/orange-cloudfoundry/custom_exporter", "-race")
	Expect(err).NotTo(HaveOccurred())

	return []byte(binaryPath)
}, func(bytes []byte) {
	binaryPath = string(bytes)
})
