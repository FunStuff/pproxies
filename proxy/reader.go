package proxy

import (
	"bufio"
	"io"
	"log"
	"regexp"
	"time"
)

var proxyRegexp = regexp.MustCompile(`((\d{1,3}\.){3}\d{1,3}):(\d+$)`)

type ProxySrc func(timeout time.Duration) ([]Proxy, error)

func HandleText(r io.Reader) ([]Proxy, error) {
	scanner := bufio.NewScanner(r)
	var str string
	var proxies []Proxy
	for scanner.Scan() {
		str = scanner.Text()
		matched := proxyRegexp.FindStringSubmatch(str)
		if len(matched) != 4 {
			continue
		}
		proxies = append(proxies, Proxy{
			HTTPS: true,
			IP:    matched[1],
			Port:  matched[3],
		})
	}
	var err error
	if err := scanner.Err(); err != nil {
		log.Println(err, "reading src reader")
	}
	return proxies, err
}
