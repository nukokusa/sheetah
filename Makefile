.PHONY: test install

test:
	go test -v ./...

install:
	go install github.com/nukokusa/sheetah/cmd/sheetah
