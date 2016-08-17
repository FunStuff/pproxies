package pproxies

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/xlaurent/pproxies/proxy"
)

type ClientList struct {
	*sync.RWMutex
	*sync.WaitGroup
	clients map[*proxyClient]struct{}
}

func NewClientList(recv <-chan proxy.Proxy, num int, timeout time.Duration, maxError int) *ClientList {
	list := &ClientList{
		&sync.RWMutex{},
		&sync.WaitGroup{},
		make(map[*proxyClient]struct{}),
	}
	for i := 0; i < num; i++ {
		client := newProxyClient(list, recv, timeout, maxError)
		list.Add(1)
		go func(client *proxyClient) {
			client.ready()
			list.Done()
		}(client)
	}
	return list
}

func (l *ClientList) add(p *proxyClient) {
	l.Lock()
	defer l.Unlock()
	l.clients[p] = struct{}{}
}

func (l *ClientList) delete(p *proxyClient) {
	l.Lock()
	defer l.Unlock()
	delete(l.clients, p)
}

type Downloader interface {
	Do(req *http.Request) (*http.Response, error)
}

func (l *ClientList) GetClient() Downloader {
	l.RLock()
	defer l.RUnlock()
	for c := range l.clients {
		return c
	}
	return http.DefaultClient
}

type proxyClient struct {
	list       *ClientList
	recv       <-chan proxy.Proxy
	client     *http.Client
	proxy      proxy.Proxy
	errCounter int
	max        int
	mutex      *sync.RWMutex
}

func newProxyClient(list *ClientList, recv <-chan proxy.Proxy, timeout time.Duration, maxError int) *proxyClient {
	return &proxyClient{
		list:   list,
		recv:   recv,
		client: &http.Client{Timeout: timeout},
		max:    maxError,
		mutex:  &sync.RWMutex{},
	}
}

func (c *proxyClient) ready() {
	p, ok := <-c.recv
	if !ok {
		return
	}
	c.proxy = p
	transport, err := p.Transport()
	if err != nil {
		return
	}
	c.client.Transport = transport
	c.list.add(c)
}

func (c *proxyClient) Do(req *http.Request) (*http.Response, error) {
	c.mutex.RLock()
	p := c.proxy
	resp, err := c.client.Do(req)
	c.mutex.RUnlock()
	if err == nil && resp.StatusCode/400 == 0 {
		return resp, err
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if p == c.proxy {
		c.errCounter++
		if c.errCounter > c.max {
			c.list.delete(c)
			c.errCounter = 0
			c.proxy = proxy.Proxy{}
			p, ok := <-c.recv
			if !ok {
				return resp, err
			}
			c.proxy = p
			transport, err := p.Transport()
			if err != nil {
				return resp, err
			}
			c.client.Transport = transport
			c.list.add(c)
		}
	}
	return resp, err
}

func (c *proxyClient) Dial() (net.Conn, error) {
	c.mutex.RLock()
	p := c.proxy
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", p.IP, p.Port))
	c.mutex.RUnlock()
	if err == nil {
		return conn, err
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if p == c.proxy {
		c.errCounter++
		if c.errCounter > c.max {
			c.list.delete(c)
			c.errCounter = 0
			c.proxy = proxy.Proxy{}
			p, ok := <-c.recv
			if !ok {
				return conn, err
			}
			c.proxy = p
			transport, err := p.Transport()
			if err != nil {
				return conn, err
			}
			c.client.Transport = transport
			c.list.add(c)
		}
	}
	return conn, err
}
