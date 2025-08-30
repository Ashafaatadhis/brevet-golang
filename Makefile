.PHONY: dev build run
 
build:
	go build -o ./bin/application.exe main.go
dev:
	air --build.cmd "go build -o .\bin\application.exe main.go" --build.bin ".\bin\application.exe"
run:
	./bin/application.exe
test:
	go test ./...