package proxy

import (
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var fakeProxies = []Proxy{
	Proxy{true, "123.45.6.7", "9999"},
	Proxy{true, "251.45.69.7", "999"},
	Proxy{true, "12.45.6.70", "99"},
	Proxy{true, "123.45.111.7", "99999"},
	Proxy{true, "1.45.6.211", "9999"},
}

func fakeApi(w http.ResponseWriter, r *http.Request) {
	for _, v := range fakeProxies {
		io.WriteString(w, v.IP+":"+v.Port+"\n")
	}
}

func TestApi(t *testing.T) {
	http.HandleFunc("/", fakeApi)
	go http.ListenAndServe(":9999", nil)
	time.Sleep(time.Second)
	proxies, err := APISrc("http://127.0.0.1:9999")(time.Second)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, fakeProxies, proxies)
}
