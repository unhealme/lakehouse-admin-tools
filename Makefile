prog    = lakehouse-admin-tools
version = debug

override LDFLAGS         := -s -w
override LDFLAGS_STATIC2 := $(LDFLAGS) -extldflags=-static
override LDFLAGS_STATIC  := $(LDFLAGS_STATIC2) -linkmode=external

.PHONY: build

build:
	go build -trimpath -ldflags='$(LDFLAGS)' -tags kerberos -o build/$(prog) ./cmd/$(prog)

dist:
	rm -f build/lakehouse-admin-tools-$(version).zip
	grep -v -e '^#' -e '^$$' lakehouse-admin-tools.conf > build/lakehouse-admin-tools.conf
	zip -j build/lakehouse-admin-tools-$(version).zip build/lakehouse-admin-tools.conf build/linux/$(prog) build/windows/$(prog).exe HOWTO.txt

static: static-linux static-windows

static-linux:
	GOOS=linux go build -trimpath -ldflags='$(LDFLAGS_STATIC)' -tags kerberos -o build/$(prog) ./cmd/$(prog)

static-windows:
	GOOS=windows go build -trimpath -ldflags='$(LDFLAGS_STATIC)' -o build/$(prog).exe ./cmd/$(prog)

release: export GOARCH = amd64
release: release-static-windows release-linux

release-windows: export GOOS = windows
release-windows:
	go build -trimpath -ldflags='$(LDFLAGS)' -o build/windows/$(prog).exe ./cmd/$(prog)
	GOAMD64=v3 go build -ldflags='$(LDFLAGS)' -o build/windows/$(prog)_amd64v3.exe ./cmd/$(prog)

release-linux: export GOOS = linux
release-linux:
	go build -trimpath -ldflags='$(LDFLAGS)' -tags kerberos -o build/linux/$(prog) ./cmd/$(prog)
	GOAMD64=v3 go build -ldflags='$(LDFLAGS)' -tags kerberos -o build/linux/$(prog)_amd64v3 ./cmd/$(prog)

release-static: export GOARCH = amd64
release-static: release-static-windows release-static-linux

release-static-windows: export GOOS = windows
release-static-windows:
	go build -trimpath -ldflags='$(LDFLAGS_STATIC2)' -o build/windows/$(prog).exe ./cmd/$(prog)
	GOAMD64=v3 go build -ldflags='$(LDFLAGS_STATIC2)' -o build/windows/$(prog)_amd64v3.exe ./cmd/$(prog)

release-static-linux: export GOOS = linux
release-static-linux:
	go build -trimpath -ldflags='$(LDFLAGS_STATIC)' -tags kerberos -o build/linux/$(prog) ./cmd/$(prog)
	GOAMD64=v3 go build -ldflags='$(LDFLAGS_STATIC)' -tags kerberos -o build/linux/$(prog)_amd64v3 ./cmd/$(prog)

clean:
	go clean -r -cache
	rm -rf build/
