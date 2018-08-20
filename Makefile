export GOOS := linux
export GOARCH := amd64

S3BUCKET := rollerbot-aws
VERSION := $(shell git describe --tags --always)

PROGS := $(subst cmd/,,$(wildcard cmd/*))

zip: $(patsubst %,dist/%.zip,$(PROGS))

bin/%: ./cmd/%/main.go internal/*.go
	go build -o $@ $<

dist/%.zip: bin/% | dist
	zip -u -j $@ $<

dist:
	mkdir dist

upload: zip
	aws s3 sync dist/ s3://$(S3BUCKET)/$(VERSION)/

.PHONY: upload
