GIT_VERSION_NAME := $(shell git describe --tags --exact-match 2> /dev/null || git symbolic-ref -q HEAD || git rev-parse HEAD)

.PHONY: csv2sqlite

csv2sqlite:
	@rm -f csv2sqlite
	@CGO_CFLAGS="-g -O3" CGO_CXXFLAGS="-g -O3" go build -ldflags "-extldflags=-static -X main.version=$(GIT_VERSION_NAME) -X main.minversion=`date -u +%Y-%m-%d.%H:%M:%S`" -tags sqlite_omit_load_extension -v -o csv2sqlite cmd/csv2sqlite/main.go
#@upx -9 csv2sqlite

sqlite-compare:
	@rm -f sqlite-compare
	@go build -o sqlite-compare cmd/sqlite-compare/main.go

check_bce:
	@CGO_CFLAGS="-g -O3" CGO_CXXFLAGS="-g -O3" go build -gcflags="-d=ssa/check_bce/debug=1" -ldflags "-extldflags=-static -X main.version=$(GIT_VERSION_NAME) -X main.minversion=`date -u +%Y-%m-%d.%H:%M:%S`" -tags sqlite_omit_load_extension -v -o csv2sqlite cmd/csv2sqlite/main.go

debug:
	@rm -f csv2sqlite_debug
	@go build -ldflags "-extldflags=-static -s -w -X main.version=$(GIT_VERSION_NAME) -X main.minversion=`date -u +%Y-%m-%d.%H:%M:%S`" -tags sqlite_omit_load_extension,debug -v -o csv2sqlite_debug cmd/csv2sqlite/main.go
	#@upx -9 csv2sqlite_debug
