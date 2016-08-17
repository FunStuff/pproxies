package pproxies

import (
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/xlaurent/pproxies/proxy"
)

const SrcTimeout = 20 * time.Second

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
	5,
	"http://httpbin.org/get",
}

func NewPool(srcs []proxy.ProxySrc) *Pool {
	return &Pool{
		srcs:   srcs,
		stopCh: make(struct{}),
		RecvCh: make(chan proxy.Proxy),
	}
}

func test(proxys []proxy.Proxy, chunkSize int, timeout time.Duration, testURL string, stop chan struct{}, out chan<- proxy.Proxy) {
	URL, err := url.Parse(testURL)
	if err != nil {
		panic(err)
	}
	https := false
	if URL.Scheme == "https" {
		https = true
	}
	chunkNum := len(proxys) / chunkSize
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
		go tester(proxys[i*chunkSize : (i+1)*chunkSize])
	}
	waiter.Add(1)
	go tester(proxys[i*chunkSize:])
	waiter.Wait()
}

func (pool *Pool) fetch(opt Option, stop chan struct{}) <-chan proxy.Proxy {
	out := make(chan proxy.Proxy, pool.chunkSize)
	waiter := &sync.WaitGroup{}
	for i, src := range pool.srcs {
		waiter.Add(1)
		go func(i int) {
			proxys, err := src(SrcTimeout)
			if err != nil {
				return
			}
			test(proxys, opt.ChunkSize, opt.Timeout, opt.TestURL, stop, out)
			waiter.Done()
		}(i)
	}
	go func() {
		waiter.Wait()
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
	var consumed bool
	for {
		if !fetching && len(buf) == 0 && consumed {
			recv = pool.fetch(opt, stop)
			fetching = true
		}
		if consumed && len(buf) != 0 {
			cache = buf[0]
			buf = buf[1:]
			send = pool.RecvCh
		}
		select {
		case p, ok := <-recv:
			if ok {
				buf = append(buf, p)
				continue
			}
			recv = nil
			fetching = false
		case send <- cache:
			send = nil
			consumed = true
		case <-stop:
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
	go pool.loop(opt, pool.stopCh)
}

func (pool *Pool) Stop() {
	select {
	case <-pool.stopCh:
		return
	default:
	}
	close(pool.stopCh)
}
