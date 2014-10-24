.PHONY: check install nuke


check:
	gofmt -w ./main.go
	go vet ./main.go

install:
	go install

nuke:
	go clean -i

build:
	rm -rf builds && mkdir builds
	go build -o builds/smcheck ./main.go
	GOOS=linux go build -o builds/smcheck_linux ./main.go
