#!/bin/bash

CGO_ENABLED=1 \
XCADDY_GO_BUILD_FLAGS="-ldflags='-w -s' -tags=nobadger,nomysql,nopgx" \
CGO_CFLAGS=$(php-config --includes) \
CGO_LDFLAGS="$(php-config --ldflags) $(php-config --libs)" \
xcaddy build \
  --output websocket \
  --with github.com/dunglas/frankenphp=./frankenphp \
  --with github.com/dunglas/frankenphp/caddy=./frankenphp/caddy \
  --with github.com/davidnurdin/frankenphp-websocket=./frankenphp-websocket \
  --with davidnurdin.com/mywebsocketserver=./davidnurdin.com/mywebsocketserver
