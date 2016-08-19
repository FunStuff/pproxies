package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/xlaurent/pproxies"
	"github.com/xlaurent/pproxies/proxy"
)

var opts struct {
	Port         int    `short:"p" json:"port" long:"port" description:"listened port" default:"9527"`
	ChunkSize    int    `short:"s" json:"chunkSize" long:"chunksize" description:"chunksize" default:"10"`
	TestURL      string `short:"u" json:"url" long:"url" description:"test url" default:"http://httpbin.org/ip"`
	TestTimeout  int    `short:"T" json:"testTimeout" long:"ttimeout" description:"test timeout(second)" default:"5"`
	ProxyTimeout int    `short:"t" json:"proxyTimeout" long:"ptimeout" description:"proxy timeout(second)" default:"5"`
	MaxError     int    `short:"e" json:"error" long:"error" description:"max error" default:"30"`
	ProxyNum     int    `short:"n" json:"proxyNum" long:"num" description:"proxy number" default:"5"`
	API          string `long:"api" json:"api" description:"proxy api" default:"default"`
	ConfigFile   string `short:"c" json:"-" long:"config" description:"config file"`
}

func HandleFlag() {
	flags.Parse(&opts)
	if opts.ConfigFile != "" {
		file, err := os.Open(opts.ConfigFile)
		if err != nil {
			log.Fatal("file not exist")
		}
		defer file.Close()
		if err := json.NewDecoder(file).Decode(&opts); err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	HandleFlag()
	var srcs []proxy.ProxySrc
	if opts.API == "default" {
		srcs = append(srcs, proxy.CyberSrc)
	} else if opts.API != "" {
		srcs = append(srcs, proxy.APISrc(opts.API))
	} else {
		log.Fatal("no API")
	}
	pool := pproxies.NewPool(srcs)
	pool.Start(&pproxies.Option{
		opts.ChunkSize,
		time.Duration(opts.TestTimeout) * time.Second,
		opts.TestURL,
	})
	recv := pool.RecvCh
	lists := pproxies.NewClientList(recv, opts.ProxyNum, time.Duration(opts.ProxyTimeout)*time.Second, opts.MaxError)
	http.ListenAndServe(":"+strconv.Itoa(opts.Port), lists)
}
