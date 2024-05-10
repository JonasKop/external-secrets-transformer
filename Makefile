all: compile-linux compile-macos
compile:
	go build -o external-secrets-transformer main.go
compile-linux:
	GOOS=linux GOARCH=amd64 go build -o external-secrets-transformer-linux-amd64 main.go
compile-macos:
	GOOS=darwin GOARCH=amd64 go build -o external-secrets-transformer-macos-amd64 main.go