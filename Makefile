BINARY := ttype
CMD := ./cmd/ttype

.PHONY: blint
build:
	go build -o bin/$(BINARY) $(CMD)

.PHONY: test
test:
	go test ./...

.PHONY: lint
lint:
	go vet ./...

.PHONY: run
run: build
	./bin/$(BINARY)

.PHONY: clean
clean:
	rm -rf bin/