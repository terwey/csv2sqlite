GIT_VERSION_NAME := $(shell git describe --tags --exact-match 2> /dev/null || git symbolic-ref -q HEAD || git rev-parse HEAD)

.PHONY: csv2sqlite

csv2sqlite:
	@rm -f csv2sqlite
	@CGO_CFLAGS="-g -O3 -DSQLITE_MAX_COLUMN=32767" CGO_CXXFLAGS="-g -O3" go build -ldflags "-extldflags=-static -X main.version=$(GIT_VERSION_NAME) -X main.minversion=`date -u +%Y-%m-%d.%H:%M:%S`" -tags sqlite_omit_load_extension -v -o csv2sqlite cmd/csv2sqlite/main.go
#@upx -9 csv2sqlite

csv2sqlite_mac:
	@rm -f csv2sqlite
	@CGO_CFLAGS="-g -O3 -DSQLITE_MAX_COLUMN=32767" CGO_CXXFLAGS="-g -O3" go build -ldflags "-extldflags=-X main.version=$(GIT_VERSION_NAME) -X main.minversion=`date -u +%Y-%m-%d.%H:%M:%S`" -tags sqlite_omit_load_extension -v -o csv2sqlite cmd/csv2sqlite/main.go

sqlite-compare:
	@rm -f sqlite-compare
	@CGO_CFLAGS="-g -O3 -DSQLITE_MAX_COLUMN=32767 -DSQLITE_EXTRA_INIT=shathree_init" CGO_CXXFLAGS="-g -O3" go build -ldflags "-extldflags=-static" -tags sqlite_omit_load_extension -v -o sqlite-compare cmd/sqlite-compare/main.go

check_bce:
	@CGO_CFLAGS="-g -O3" CGO_CXXFLAGS="-g -O3" go build -gcflags="-d=ssa/check_bce/debug=1" -ldflags "-extldflags=-static -X main.version=$(GIT_VERSION_NAME) -X main.minversion=`date -u +%Y-%m-%d.%H:%M:%S`" -tags sqlite_omit_load_extension -v -o csv2sqlite cmd/csv2sqlite/main.go

debug:
	@rm -f csv2sqlite_debug
	@go build -ldflags "-extldflags=-static -s -w -X main.version=$(GIT_VERSION_NAME) -X main.minversion=`date -u +%Y-%m-%d.%H:%M:%S`" -tags sqlite_omit_load_extension,debug -v -o csv2sqlite_debug cmd/csv2sqlite/main.go
	#@upx -9 csv2sqlite_debug

markdown:
	@asciidoctor docs/readme.adoc -b docbook -a leveloffset=+1 -o - | pandoc  --atx-headers --wrap=preserve -t gfm -f docbook - > README.md