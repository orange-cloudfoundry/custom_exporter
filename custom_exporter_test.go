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

	"github.com/alicebob/miniredis"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
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

type failRunner struct {
	Command           *exec.Cmd
	Name              string
	AnsiColorCode     string
	StartCheck        string
	StartCheckTimeout time.Duration
	Cleanup           func()
	session           *gexec.Session
	sessionReady      chan struct{}
	existStatus       int
}

var (
	args        []string
	listenAddr  string
	metricRoute string
	configPath  string
	logLevel    string

	server  *miniredis.Miniredis
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
			Expect(session.ExitCode()).To(Equal(r.existStatus), fmt.Sprintf("Expected process to exit with %d, got: %d", r.existStatus, session.ExitCode()))
			return nil
		}
	}
}

var _ = Describe("Custom Export Main Test", func() {
	BeforeEach(func() {
		logLevel = "debug"
	})

	AfterEach(func() {
		ginkgomon.Kill(process)
	})

	Context("Missing required args", func() {
		It("shows usage", func() {
			var args []string

			args = append(args, "-log.level="+logLevel)

			exporter := failRunner{
				Name:        "custom_exporter",
				Command:     exec.Command(binaryPath, args...),
				StartCheck:  " missing required -collector.config argument/flag",
				existStatus: 2,
			}
			process = ifrit.Invoke(exporter)
		})
	})

	Context("Given a wrong required args", func() {
		It("shows usage", func() {
			var args []string

			args = append(args, "-collector.config=wrong.err")
			args = append(args, "-log.level="+logLevel)

			exporter := failRunner{
				Name:        "custom_exporter",
				Command:     exec.Command(binaryPath, args...),
				StartCheck:  "no such file or directory",
				existStatus: 2,
			}

			process = ifrit.Invoke(exporter)
		})
	})

	Context("Has required args", func() {
		BeforeEach(func() {
			listenAddr = "0.0.0.0:" + strconv.Itoa(9213+GinkgoParallelNode())
			configPath = "example_shell.yml"
			metricRoute = "/metrics"

			args = append(args, "-web.listen-address="+listenAddr)
			args = append(args, "-collector.config="+configPath)
			args = append(args, "-web.telemetry-path="+metricRoute)
			args = append(args, "-log.level="+logLevel)

			exporter := failRunner{
				Name:              "custom_exporter",
				Command:           exec.Command(binaryPath, args...),
				StartCheck:        "Listening",
				StartCheckTimeout: 30 * time.Second,
				existStatus:       137,
			}

			process = ifrit.Invoke(exporter)
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

			//println(string(body))

			Expect(string(body)).To(ContainSubstring("custom_custom_metric_shell{animals=\"beef\",id=\"2\"} 256"))
			Expect(string(body)).To(ContainSubstring("custom_custom_metric_shell{animals=\"chicken\",id=\"1\"} 128"))
			Expect(string(body)).To(ContainSubstring("custom_custom_metric_shell{animals=\"snails\",id=\"3\"} 14"))
		})

	})
})
