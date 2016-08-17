package proxy

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/xlaurent/calc"
)

func CyberSrc(timeout time.Duration) ([]Proxy, error) {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get("http://www.cybersyndrome.net/search.cgi?q=CN&a=ABC&f=s&s=new&n=500")
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, err
	}
	page := doc.Find("#content > script").Text()
	return extractProxies(page), nil
}

func extractProxies(page string) []Proxy {
	arraysExp := regexp.MustCompile(`\[(.+?)\]`)
	matched := arraysExp.FindAllStringSubmatch(page, 2)
	as := strings.Split(matched[0][1], ",")
	ps := strings.Split(matched[1][1], ",")

	computeExp := regexp.MustCompile(`(\(.+\)%\d+)`)
	expr := computeExp.FindStringSubmatch(page)[1]
	expr = strings.Replace(expr, "[", "(", -1)
	expr = strings.Replace(expr, "]", ")", -1)

	calc.RegisterStringSlice(as, "as")
	calc.RegisterStringSlice(ps, "ps")

	n, err := calc.Evaluate(expr)
	if err != nil {
		return nil
	}

	as = append(as[int(n):], as[0:int(n)]...)
	var proxies []Proxy
	for i := range as {
		idx := i / 4
		if i%4 == 0 {
			proxies = append(proxies, Proxy{true, as[i] + ".", ps[idx]})
		} else if i%4 == 3 {
			proxies[idx].IP += as[i]
		} else {
			proxies[idx].IP += as[i] + "."
		}
	}
	return proxies
}
