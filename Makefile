GIT_VERSION_NAME := $(shell git describe --tags --exact-match 2> /dev/null || git symbolic-ref -q HEAD || git rev-parse HEAD)

.PHONY: csv2sqlite

csv2sqlite:
	@rm -f csv2sqlite
	@go build -ldflags "-extldflags=-static -s -w -X main.version=$(GIT_VERSION_NAME) -X main.minversion=`date -u +%Y-%m-%d.%H:%M:%S`" -tags sqlite_omit_load_extension -i -v -o csv2sqlite cmd/csv2sqlite/main.go
	@upx -9 csv2sqlite