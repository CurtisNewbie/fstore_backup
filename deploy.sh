#!/bin/bash

os="linux"
arch="amd64"
build="fbackuprun"
CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -o "$build"

[[ $? -eq 0 ]] && scp fbackuprun alphaboi@curtisnewbie.com:/home/alphaboi/services/fstore_backup
[[ -f "$build" ]] && rm "$build"
