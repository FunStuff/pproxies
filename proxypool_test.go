package pproxies

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/xlaurent/pproxies/proxy"
	"github.com/xlaurent/pproxies/testutil"
)

const (
	serverPort   = 9999
	proxyStart   = 8000
	proxyNum     = 20
	srcChunkSize = 10
)

var state int = 0

func fakeSrc(timeout time.Duration) ([]proxy.Proxy, error) {
	var proxies []proxy.Proxy
	for j := state * srcChunkSize; j < (state+1)*srcChunkSize; j++ {
		proxies = append(proxies, proxy.Proxy{
			true,
			"127.0.0.1",
			strconv.Itoa(proxyStart + j),
		})
	}
	if state == proxyNum/srcChunkSize {
		state = 0
	} else {
		state++
	}
	return proxies, nil
}

func TestMain(m *testing.M) {
	for i := 0; i < proxyNum; i++ {
		go testutil.RunHTTPProxy(i + proxyStart)
	}
	go testutil.TestServer(serverPort)
	os.Exit(m.Run())
}

func TestPool(t *testing.T) {
	pool := NewPool([]proxy.ProxySrc{fakeSrc})
	pool.Start(&Option{
		6,
		5 * time.Second,
		fmt.Sprintf("http://127.0.0.1:%d", serverPort),
	})
	recv := pool.RecvCh
	var proxies []proxy.Proxy
	for i := 0; i < proxyNum; i++ {
		p := <-recv
		proxies = append(proxies, p)
	}
	assert.Len(t, proxies, 20)
	pool.Stop()
	select {
	case <-recv:
	case <-time.After(10 * time.Second):
		t.Error("can't stop\n")
	}
}
