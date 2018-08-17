export GOOS := linux
export GOARCH := amd64

S3BUCKET := rollerbot-aws
VERSION := $(git describe --tags --always)
PROGS := \
	count-outdated-instances \
	start-roller

zip: $(patsubst %,dist/%.zip,$(PROGS))

bin/%: ./cmd/%/main.go
	go build -o $@ $<

dist/%.zip: bin/% | dist
	zip -u -j $@ $<

dist:
	mkdir dist

upload: zip
	aws s3 sync dist/*.zip s3://$(S3BUCKET)/$(VERSION)/

.PHONY: upload
