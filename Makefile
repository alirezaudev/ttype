BINARY := ttype
CMD := ./cmd/ttype

.PHONY: build
build:
	go build -o bin/$(BINARY) $(CMD)

.PHONY: test
test:
	go test ./...

.PHONY: cover
cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

.PHONY: cover-html
cover-html: cover
	go tool cover -html=coverage.out

.PHONY: lint
lint:
	go vet ./...

.PHONY: run
run: build
	./bin/$(BINARY)

.PHONY: clean
clean:
	rm coverage.out
	rm -rf bin/