#!/bin/sh
go get github.com/Khan/genqlient/generate
go run github.com/Khan/genqlient
go mod tidy
