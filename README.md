[![Go Report Card](https://goreportcard.com/badge/github.com/thehapax/lnurlpayserver)](https://goreportcard.com/report/github.com/thehapax/lnurlpayserver)

# README for lnurlpayserver

## How to Install

Requirements: 
on OSX 10.13:

Install Go version 1.13.8
```
$ brew update
$ brew install golang
$ brew install postgresql
$ brew services start postgresql
```
if you have an earlier version of goland you can change.
```
$ brew switch go 1.13.8
```

setup postgres db
```
$ psql
postgres=# createdb `lnurlpaydb`
postgres=# createuser -s postgres

$ psql -U postgres -h 127.0.0.1 -d lnurlpaydb -f postgres.sql
```

Edit your ~/.bash_profile accordingly:
```
export GOPATH=$HOME/go-workspace # don't forget to change your path correctly!
export GOROOT=/usr/local/opt/go/libexec
export PATH=$PATH:$GOPATH/bin
export PATH=$PATH:$GOROOT/bin
```
source your profile:
```
$ source ~/.bash_profile
```

install all go dependencies and sub dependencies:
```
$ go get -u -v -f all

$ npm install 

$ make lnurlpayserver
```

Setup your environment variable in a .env, for example:
```
#!/bin/bash

export HOST=localhost
export PORT=2000
export SERVICE_URL=https://yourdomain.com
export DATABASE_URL=postgres://user:password@host:port/lnurlpaydb
export SECRET=anything_here
```

start the server 
```
./lnurlpayserver 
```

or if you have godotenv, 
```
godotenv -f .env ./lnurlpayserver 
```

## Todo

Write instructions on the following:
- [ ] How to Setup a Shop to a lightning backend in the Web interface
- [ ] How to create a template (which is like an item from the shop
- [ ] How to define infinte advance and complex parameters
- [ ] How to add items and generate an invoice
- [ ] How to generate the API

## Troubleshooting

- SSL: If you want to run on localhost with no SSL, you'll need to append to the DATABASE_URL to disable ssl, example:

```
 export DATABASE_URL=postgres://username:password@localhost:5432/lnurlpay?sslmode=disable
```

- If you get a blank page when go to the web: 
  - First, check that the public directory is not empty
  - If is empty, that means preact wasn't built so do `make public` to build it


## Known dependencies for this project:

```
$ go get -u <<dependencies>>

 github.com/go-bindata/go-bindata
 github.com/itchyny/gojq
 github.com/fiatjaf/go-lnurl
 github.com/fiatjaf/lightningd-gjson-rpc
 github.com/fiatjaf/ln-decodepay
 github.com/fiatjaf/lunatico
 github.com/gorilla/mux
 github.com/hoisie/mustache
 github.com/jmoiron/sqlx
 github.com/jmoiron/sqlx/types
 github.com/kelseyhightower/envconfig
 github.com/lib/pq
 github.com/orcaman/concurrent-map
 github.com/rs/cors
 github.com/rs/zerolog
 github.com/tidwall/gjson
 github.com/tidwall/sjson
```
