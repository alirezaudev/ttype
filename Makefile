BINARY := ttype
CMD := ./cmd/ttype
PREFIX ?= /usr/local
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
TAPES := $(patsubst docs/tapes/%.tape,%,$(wildcard docs/tapes/[!_]*.tape))

.PHONY: build
build:
	go build -ldflags "-X main.version=$(VERSION)" -o bin/$(BINARY) $(CMD)

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

.PHONY: demo
# Record docs/tapes/*.tape into docs/*.gif. `make demo TAPE=modes` for one.
# Tapes run from the repo root and point PATH at ./bin, and they isolate the
# config and data dirs so recording never touches real history — which only
# works on Linux.
demo: build
	@command -v vhs >/dev/null || { echo "vhs not installed: https://github.com/charmbracelet/vhs"; exit 1; }
	@for t in $(if $(TAPE),$(TAPE),$(TAPES)); do \
		test -f docs/tapes/$$t.tape || { echo "no such tape: $$t"; exit 1; }; \
		echo "recording $$t"; \
		vhs docs/tapes/$$t.tape || exit 1; \
	done

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

.PHONY: release-snapshot
release-snapshot:
	goreleaser release --snapshot --clean
