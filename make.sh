#!/usr/bin/env sh

export CGO_ENABLED=0
go build -v -o ipapi-agent .
