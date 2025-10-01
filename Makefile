install-dependencies:
	echo "Installing dependencies..."

generate:
	go generate ./...

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8 run ./...
	
fix:
	golangci-lint run --fix

test:
	go test ./... --cover

