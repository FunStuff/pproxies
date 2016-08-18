pproxies-alpine:pproxies.go
	docker run -v $$GOPATH:/go --rm golang:alpine /bin/sh -c "cd /go/src/github.com/xlaurent/pproxies/cmd/ && go build -o pproxies-alpine"

image:pproxies-alpine
	docker build -t xlaurent/pproxies .
