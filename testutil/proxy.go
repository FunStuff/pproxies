package testutil

import (
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"
)

func RunHTTPProxy(port int) {
	http.ListenAndServe(":"+strconv.Itoa(port), http.HandlerFunc(handleFunc(port)))
}

func handleFunc(port int) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "CONNECT" {
			hj, _ := w.(http.Hijacker)
			cConn, bufrw, err := hj.Hijack()
			defer cConn.Close()
			if err != nil {
				bufrw.WriteString("HTTP/1.1 502 Bad Gateway")
				bufrw.Flush()
				return
			}
			sConn, err := net.Dial("tcp", r.Host)
			defer sConn.Close()
			if err != nil {
				bufrw.WriteString("HTTP/1.1 502 Bad Gateway")
				bufrw.Flush()
				return
			}
			bufrw.WriteString("HTTP/1.1 200 Connection Established\r\n\r\n")
			bufrw.Flush()
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
		req, err := http.NewRequest(r.Method, r.URL.String(), r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		req.Header["proxy"] = []string{strconv.Itoa(port)}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		w.WriteHeader(resp.StatusCode)
		header := w.Header()
		for k, v := range resp.Header {
			header[k] = v
		}
		io.Copy(w, resp.Body)
		return
	}
}
