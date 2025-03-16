#!/bin/sh
set -x

src=proto/custom
dst=proto/gen

# ensure target dir exists
mkdir -p $dst

protoc -I proto -I proto/import \
  --go_opt=paths=source_relative --go_out=$dst \
  $src/*.proto