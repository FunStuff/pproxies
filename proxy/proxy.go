package proxy

import (
	"fmt"
	"net/http"
	"net/url"
)

type Proxy struct {
	HTTPS bool
	IP    string
	Port  string
}

func (p Proxy) Test(client *http.Client, URL string) error {
	transport, err := p.Transport()
	if err != nil {
		return err
	}
	client.Transport = transport
	resp, err := client.Get(URL)
	if err != nil {
		return err
	}
	if resp.StatusCode/400 > 0 {
		return fmt.Errorf("status code:%d", resp.StatusCode)
	}
	return nil
}

func (p Proxy) Transport() (*http.Transport, error) {
	URL, err := url.Parse(p.String())
	if err != nil {
		return nil, fmt.Errorf("can't parse proxy url %s,%v", p.String(), err)
	}
	return &http.Transport{
		Proxy: http.ProxyURL(URL),
	}, nil
}

func (p Proxy) String() string {
	return fmt.Sprintf("http://%s:%s", p.IP, p.Port)
}
