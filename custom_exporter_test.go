package main_test

import (
	"io"
	"net/http"
	"os/exec"
	"strconv"

	"fmt"

	"os"
	"time"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/prometheus/common/log"
	"io/ioutil"
)

type failRunner struct {
	Command           *exec.Cmd
	Name              string
	AnsiColorCode     string
	StartCheck        string
	StartCheckTimeout time.Duration
	Cleanup           func()
	session           *gexec.Session
	sessionReady      chan struct{}
}

var (
	args        []string
	listenAddr  string
	metricRoute string
	configPath  string

	process ifrit.Process
)

func (r failRunner) Run(sigChan <-chan os.Signal, ready chan<- struct{}) error {
	defer GinkgoRecover()

	allOutput := gbytes.NewBuffer()

	debugWriter := gexec.NewPrefixedWriter(
		fmt.Sprintf("\x1b[32m[d]\x1b[%s[%s]\x1b[0m ", r.AnsiColorCode, r.Name),
		GinkgoWriter,
	)

	session, err := gexec.Start(
		r.Command,
		gexec.NewPrefixedWriter(
			fmt.Sprintf("\x1b[32m[o]\x1b[%s[%s]\x1b[0m ", r.AnsiColorCode, r.Name),
			io.MultiWriter(allOutput, GinkgoWriter),
		),
		gexec.NewPrefixedWriter(
			fmt.Sprintf("\x1b[91m[e]\x1b[%s[%s]\x1b[0m ", r.AnsiColorCode, r.Name),
			io.MultiWriter(allOutput, GinkgoWriter),
		),
	)

	Ω(err).ShouldNot(HaveOccurred())

	fmt.Fprintf(debugWriter, "spawned %s (pid: %d)\n", r.Command.Path, r.Command.Process.Pid)

	r.session = session
	if r.sessionReady != nil {
		close(r.sessionReady)
	}

	startCheckDuration := r.StartCheckTimeout
	if startCheckDuration == 0 {
		startCheckDuration = 5 * time.Second
	}

	var startCheckTimeout <-chan time.Time
	if r.StartCheck != "" {
		startCheckTimeout = time.After(startCheckDuration)
	}

	detectStartCheck := allOutput.Detect(r.StartCheck)

	for {
		select {
		case <-detectStartCheck: // works even with empty string
			allOutput.CancelDetects()
			startCheckTimeout = nil
			detectStartCheck = nil
			close(ready)

		case <-startCheckTimeout:
			// clean up hanging process
			session.Kill().Wait()

			// fail to start
			return fmt.Errorf(
				"did not see %s in command's output within %s. full output:\n\n%s",
				r.StartCheck,
				startCheckDuration,
				string(allOutput.Contents()),
			)

		case signal := <-sigChan:
			session.Signal(signal)

		case <-session.Exited:
			if r.Cleanup != nil {
				r.Cleanup()
			}

			Expect(string(allOutput.Contents())).To(ContainSubstring(r.StartCheck))
			Expect(session.ExitCode()).To(Equal(1), fmt.Sprintf("Expected process to exit with 1, got: %d", session.ExitCode()))
			return nil
		}
	}
}

var _ = Describe("Custom Export Main Test", func() {
	Context("Missing required args", func() {
		It("shows usage", func() {
			var args []string
			exporter := failRunner{
				Name:       "customExporter",
				Command:    exec.Command(binaryPath, args...),
				StartCheck: "Config file parameter must be provided",
			}
			process := ifrit.Invoke(exporter)
			ginkgomon.Kill(process) // this is only if incorrect implementation leaves process running
		})
	})

	Context("Given a wrong required args", func() {
		It("shows usage", func() {
			var args []string

			args = append(args, "-collector.config=wrong.err")

			exporter := failRunner{
				Name:       "customExporter",
				Command:    exec.Command(binaryPath, args...),
				StartCheck: "no such file or directory",
			}

			process := ifrit.Invoke(exporter)
			ginkgomon.Kill(process) // this is only if incorrect implementation leaves process running
		})
	})

	Context("Has required args", func() {
		BeforeEach(func() {
			listenAddr = "0.0.0.0:" + strconv.Itoa(9209+GinkgoParallelNode())
			configPath = "example.yml"
			metricRoute = "/metrics"

			args = append(args, "-web.listen-address="+listenAddr)
			args = append(args, "-collector.config="+configPath)
			args = append(args, "-web.telemetry-path="+metricRoute)

			exporterCustom := ginkgomon.New(ginkgomon.Config{
				Name:              "customExporter",
				Command:           exec.Command(binaryPath, args...),
				StartCheck:        "Listening",
				StartCheckTimeout: 30 * time.Second,
			})

			process = ginkgomon.Invoke(exporterCustom)
		})

		AfterEach(func() {
			ginkgomon.Kill(process)
		})

		It("should listen on the given address and return the landing page", func() {

			landingPage := []byte(`<html><head><title>Custom exporter</title></head><body><h1>Custom exporter</h1><p><a href='/metrics'>Metrics</a></p></body></html>`)

			req, err := http.NewRequest("GET", "http://"+listenAddr+"/", nil)
			Expect(err).NotTo(HaveOccurred())

			resp, err := http.DefaultClient.Do(req)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))

			body, err := ioutil.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(Equal(landingPage))
		})

		It("should listen on the given address and return the metrics route", func() {

			req, err := http.NewRequest("GET", "http://"+listenAddr+"/"+metricRoute, nil)
			Expect(err).NotTo(HaveOccurred())

			resp, err := http.DefaultClient.Do(req)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))

			body, err := ioutil.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())

			log.Infoln(string(body))
		})

	})
})
