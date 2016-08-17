package proxy

import (
	"net/http"
	"time"
)

func APISrc(url string) ProxySrc {
	return func(timeout time.Duration) ([]Proxy, error) {
		client := http.Client{Timeout: timeout}
		resp, err := client.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		return HandleText(resp.Body)
	}
}
