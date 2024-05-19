build:
	 @cd src && go build -o ../bin/gobank

run: build
	@./bin/gobank

test: 
	@go test -v ./...