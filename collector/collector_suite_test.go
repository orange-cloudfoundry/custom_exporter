package collector_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alicebob/miniredis"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"testing"
)

var (
	redisAddr   string
	redisServer *miniredis.Miniredis
	ch          chan prometheus.Metric
	ds          chan *prometheus.Desc
)

func TestCustomExporter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Custom Config Test Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	var err error

	if redisServer, err = miniredis.Run(); err != nil {
		println(err.Error())
		panic(err)
	}

	redisAddr = redisServer.Addr()
	log.Infof("Miniredis started and listinning on Addr \"%s\" ...", redisAddr)

	return []byte(redisAddr)
}, func(byte []byte) {
	ch = make(chan prometheus.Metric)
	ds = make(chan *prometheus.Desc)
	log.Infoln("Channels openned...")

	redisServer.FlushAll()
	redisServer.RequireAuth("password")
	redisServer.Set("foo1", "{\"test\":1,\"role\":\"master\",\"value\":\"14.258\"}")
	redisServer.Set("foo2", "{\"test\":2,\"role\":\"master\",\"value\":\"6843.119\"}")
	redisServer.Set("foo3", "{\"test\":3,\"role\":\"master\",\"value\":\"18.1244\"}")
	redisServer.Set("foo4", "{\"test\":4,\"role\":\"master\",\"value\":\"15.2234841e+12\"}")
})

var _ = SynchronizedAfterSuite(func() {
	log.Infof("Stopping Miniredis listinning on Addr \"%s\" ...", redisAddr)
	redisServer.Close()

}, func() {
	log.Infoln("Closing Channels...")
	close(ch)
	close(ds)
})
