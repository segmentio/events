test:
	go vet ./...
	go test -race -v ./...

vendor:
	go mod vendor