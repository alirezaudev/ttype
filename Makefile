BINARY := ttype
CMD := ./cmd/ttype
PREFIX ?= /usr/local

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

.PHONY: lint-strict
lint-strict:
	golangci-lint run ./...

.PHONY: bench
bench:
	go test -run XXX -bench . -benchmem ./...

.PHONY: run
run: build
	./bin/$(BINARY)

.PHONY: clean
clean:
	rm coverage.out
	rm -rf bin/
.PHONY: completions
completions: build
	@mkdir -p contrib/completions
	./bin/$(BINARY) completion bash       > contrib/completions/$(BINARY).bash
	./bin/$(BINARY) completion zsh        > contrib/completions/_$(BINARY)
	./bin/$(BINARY) completion fish       > contrib/completions/$(BINARY).fish
	./bin/$(BINARY) completion powershell > contrib/completions/$(BINARY).ps1

.PHONY: man-install
man-install:
	install -Dm644 man/$(BINARY).1 $(DESTDIR)$(PREFIX)/share/man/man1/$(BINARY).1
