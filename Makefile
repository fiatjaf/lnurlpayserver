lnurlpayserver: $(shell find . -name "*.go") bindata.go
	go build

public: $(shell find ./client)
	mkdir -p public
	./node_modules/.bin/preact build --src ./client --dest ./public/ --no-prerender

bindata.go: public
	go-bindata -o bindata.go public/...
