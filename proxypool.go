package pproxies

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/xlaurent/pproxies/proxy"
)

const SrcTimeout = 20 * time.Second

var logger *log.Logger

func init() {
	logger = log.New(os.Stderr, "info:", log.LstdFlags)
}

type Pool struct {
	srcs   []proxy.ProxySrc
	stopCh chan struct{}
	RecvCh chan proxy.Proxy
}

type Option struct {
	ChunkSize int
	Timeout   time.Duration
	TestURL   string
}

var defaultOpt = &Option{
	6,
	5 * time.Second,
	"http://httpbin.org/ip",
}

func NewPool(srcs []proxy.ProxySrc) *Pool {
	ch := make(chan struct{})
	close(ch)
	return &Pool{
		srcs:   srcs,
		stopCh: ch,
	}
}

func test(proxies []proxy.Proxy, chunkSize int, timeout time.Duration, testURL string, stop chan struct{}, out chan<- proxy.Proxy) {
	URL, err := url.Parse(testURL)
	if err != nil {
		panic(err)
	}
	https := false
	if URL.Scheme == "https" {
		https = true
	}
	chunkNum := len(proxies) / chunkSize
	waiter := &sync.WaitGroup{}
	tester := func(ps []proxy.Proxy) {
		defer waiter.Done()
		client := &http.Client{Timeout: timeout}
		for _, p := range ps {
			if https && !p.HTTPS {
			} else if err := p.Test(client, testURL); err != nil {
			} else {
				select {
				case <-stop:
					return
				case out <- p:
				}
				continue
			}
			select {
			case <-stop:
				return
			default:
			}
		}
	}
	var i int
	for i = 0; i < chunkNum-1; i++ {
		waiter.Add(1)
		go tester(proxies[i*chunkSize : (i+1)*chunkSize])
	}
	waiter.Add(1)
	go tester(proxies[i*chunkSize:])
	waiter.Wait()
}

func (pool *Pool) fetch(opt Option, stop chan struct{}) <-chan proxy.Proxy {
	out := make(chan proxy.Proxy, opt.ChunkSize)
	waiter := &sync.WaitGroup{}
	for i, src := range pool.srcs {
		waiter.Add(1)
		go func(i int) {
			defer waiter.Done()
			proxies, err := src(SrcTimeout)
			if err != nil {
				return
			}
			logger.Printf("fetched %d proxies from src %d\n", len(proxies), i)
			test(proxies, opt.ChunkSize, opt.Timeout, opt.TestURL, stop, out)
		}(i)
	}
	go func() {
		waiter.Wait()
		logger.Println("finish testing")
		close(out)
	}()
	return out
}

func (pool *Pool) loop(opt Option, stop chan struct{}) {
	var recv <-chan proxy.Proxy
	var send chan<- proxy.Proxy
	var buf []proxy.Proxy
	var cache proxy.Proxy
	var fetching bool
	var consumed = true
	var count int
	for {
		if !fetching && len(buf) == 0 && consumed {
			logger.Println("start fetching proxies")
			recv = pool.fetch(opt, stop)
			fetching = true
		}
		if consumed && len(buf) != 0 {
			consumed = false
			cache = buf[0]
			buf = buf[1:]
			send = pool.RecvCh
		}
		select {
		case p, ok := <-recv:
			if ok {
				count++
				buf = append(buf, p)
				continue
			}
			logger.Printf("a total of %d available proxies\n", count)
			count = 0
			recv = nil
			fetching = false
		case send <- cache:
			send = nil
			consumed = true
		case <-stop:
			close(pool.RecvCh)
			return
		}
	}
}

func (pool *Pool) Start(opt *Option) {
	select {
	case <-pool.stopCh:
	default:
		return
	}
	if opt == nil {
		opt = defaultOpt
	}
	pool.stopCh = make(chan struct{})
	pool.RecvCh = make(chan proxy.Proxy)
	go pool.loop(*opt, pool.stopCh)
}

func (pool *Pool) Stop() {
	select {
	case <-pool.stopCh:
		return
	default:
	}
	close(pool.stopCh)
}
