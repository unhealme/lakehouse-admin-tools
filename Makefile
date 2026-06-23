prog = lakehouse-admin-tools

build:
	go build -o build/$(prog)

static:
	go build -trimpath -ldflags="-extldflags=-static" -o build/$(prog)

release: export GOARCH = amd64
release:
	GOOS=windows go build -a -trimpath -ldflags="-s -w -extldflags=-static" -o build/windows/$(prog).exe
	GOOS=linux go build -a -trimpath -ldflags="-s -w -extldflags=-static" -o build/linux/$(prog)
	GOAMD64=v3 GOOS=windows go build -a -trimpath -ldflags="-s -w -extldflags=-static" -o build/windows/$(prog)_amd64v3.exe
	GOAMD64=v3 GOOS=linux go build -a -trimpath -ldflags="-s -w -extldflags=-static" -o build/linux/$(prog)_amd64v3

clean:
	go clean -r -cache
	rm -rf build/
