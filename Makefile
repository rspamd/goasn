goasn: clean vendor patch build

build:
	CGO_ENABLED=0 go build

clean:
	rm -rf vendor

vendor:
	go mod vendor

patch:
	git apply gobgp.patch
