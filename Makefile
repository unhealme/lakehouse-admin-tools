prog = lakehouse-admin-tools
version = debug

build:
	go build -o build/$(prog)

dist:
	rm -f build/lakehouse-admin-tools-$(version).zip
	grep -v -e '^#' -e '^$$' lakehouse-admin-tools.conf > build/lakehouse-admin-tools.conf
	zip -j build/lakehouse-admin-tools-$(version).zip build/lakehouse-admin-tools.conf build/linux/$(prog) build/windows/$(prog).exe HOWTO.txt

static: static-linux static-windows

static-linux:
	GOOS=linux go build -trimpath -ldflags="-s -w -extldflags=-static" -o build/$(prog)

static-windows:
	GOOS=windows go build -trimpath -ldflags="-s -w -extldflags=-static" -o build/$(prog).exe

release: export GOARCH = amd64
release:
	@echo GOARCH=$$GOARCH
	GOOS=windows go build -trimpath -ldflags="-s -w -extldflags=-static" -o build/windows/$(prog).exe
	GOOS=linux go build -trimpath -ldflags="-s -w -extldflags=-static" -o build/linux/$(prog)
	GOAMD64=v3 GOOS=windows go build -trimpath -ldflags="-s -w -extldflags=-static" -o build/windows/$(prog)_amd64v3.exe
	GOAMD64=v3 GOOS=linux go build -trimpath -ldflags="-s -w -extldflags=-static" -o build/linux/$(prog)_amd64v3

clean:
	go clean -r -cache
	rm -rf build/
