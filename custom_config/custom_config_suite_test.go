package custom_config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

var binaryPath string

func TestCustomExporter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Custom Config Test Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	var err error
	binaryPath, err = gexec.Build("github.com/orange-cloudfoundry/custom_exporter", "-race")
	Expect(err).NotTo(HaveOccurred())

	return []byte(binaryPath)
}, func(bytes []byte) {
	binaryPath = string(bytes)
})
