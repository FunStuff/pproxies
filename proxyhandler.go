package pproxies

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
)

func writeError(bufrw *bufio.ReadWriter) {
	bufrw.WriteString("HTTP/1.1 502 Bad Gateway")
	bufrw.Flush()
}

func (h *ClientList) handleHTTPS(w http.ResponseWriter, r *http.Request) {
	hj, _ := w.(http.Hijacker)
	cConn, bufrw, err := hj.Hijack()
	defer cConn.Close()
	if err != nil {
		writeError(bufrw)
		return
	}
	dler, ok := h.GetClient().(*proxyClient)
	var sConn net.Conn
	if !ok {
		sConn, err = net.Dial("tcp", r.Host)
		if err != nil {
			writeError(bufrw)
			return
		}
		defer sConn.Close()
		bufrw.WriteString("HTTP/1.1 200 Connection Established\r\n\r\n")
		bufrw.Flush()
	} else {
		sConn, err = dler.Dial()
		if err != nil {
			writeError(bufrw)
			return
		}
		_, err := io.WriteString(sConn, fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", r.Host, r.Host))
		if err != nil {
			writeError(bufrw)
			return
		}
		defer sConn.Close()
	}
	waiter := &sync.WaitGroup{}
	waiter.Add(1)
	go func() {
		io.Copy(sConn, bufrw)
		waiter.Done()
	}()
	waiter.Add(1)
	go func() {
		io.Copy(cConn, sConn)
		waiter.Done()
	}()
	waiter.Wait()
	return
}

func (h *ClientList) handleHTTP(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest(r.Method, r.URL.String(), r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	req.Header = r.Header
	client := h.GetClient()
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	header := w.Header()
	for k, v := range resp.Header {
		header[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
	return
}

func (h *ClientList) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "CONNECT" {
		h.handleHTTPS(w, r)
		return
	}
	h.handleHTTP(w, r)
}
