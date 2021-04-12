#!/usr/bin/env bash

set -e

go test -v -coverpkg=$1 -coverprofile=profile.out $1
go tool cover -func profile.out
