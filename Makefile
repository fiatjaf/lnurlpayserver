lnurlpayserver: $(shell find . -name "*.go") bindata.go
	go build

public: $(shell find ./client)
	mkdir -p public
	./node_modules/.bin/preact build --src ./client --dest ./public/ --no-prerender

bindata.go: public
	go-bindata -o bindata.go public/...

deploy: lnurlpayserver
	ssh root@nusakan-58 'systemctl stop lnurlpayserver'
	scp lnurlpayserver nusakan-58:lnurlpayserver/lnurlpayserver
	ssh root@nusakan-58 'systemctl start lnurlpayserver'
