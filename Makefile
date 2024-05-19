build:
	 @cd src && go build -o ../bin/j-bank

run: build
	@./bin/j-bank

test: 
	@go test -v ./...