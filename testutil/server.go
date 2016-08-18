package testutil

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type Count struct {
	first time.Time
	count int64
}

func TestServer(port int) {
	m := make(map[string]*Count)
	ban := make(map[string]struct{})
	mutex := &sync.Mutex{}
	banMutex := &sync.RWMutex{}
	handler := http.NewServeMux()
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		pport, ok := r.Header["Proxy"]
		if !ok {
			w.WriteHeader(403)
			return
		}
		port := pport[0]
		banMutex.RLock()
		if _, ok := ban[port]; ok {
			banMutex.RUnlock()
			w.WriteHeader(403)
			return
		}
		if len(ban) == 20 {
			for k, v := range m {
				fmt.Println(k, v.count)
			}
		}
		os.Exit(0)
		banMutex.RUnlock()
		mutex.Lock()
		if cnt, ok := m[port]; !ok {
			m[port] = &Count{
				first: time.Now(),
				count: 1,
			}
		} else {
			ps := cnt.count / int64(time.Now().Sub(cnt.first)) / int64(time.Second)
			if ps < 100 {
				m[port].count = m[port].count + 1
				mutex.Unlock()
			} else {
				mutex.Unlock()
				banMutex.Lock()
				ban[port] = struct{}{}
				banMutex.Unlock()
			}
		}
		w.WriteHeader(202)
	})
	http.ListenAndServe(":"+strconv.Itoa(port), handler)
}
