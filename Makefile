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
# Record docs/tapes/*.tape into docs/*.svg. `make demo TAPE=modes` for one.
# Each tape is played in a pty and the recording is turned into an animated
# SVG. The recorder isolates the config and data dirs so recording never
# touches real history, which only works on Linux.
demo: build
	@python3 -c "import pyte" 2>/dev/null || { echo "pyte not installed: pip install pyte"; exit 1; }
	@for t in $(if $(TAPE),$(TAPE),$(TAPES)); do \
		test -f docs/tapes/$$t.tape || { echo "no such tape: $$t"; exit 1; }; \
		case $$t in demo) size="80 22";; help) size="96 40";; *) size="96 28";; esac; \
		cast=$$(mktemp); \
		python3 docs/tapes/tape2cast.py docs/tapes/$$t.tape $$cast $$size $(CURDIR)/bin && \
		python3 docs/tapes/cast2svg.py $$cast docs/$$t.svg || exit 1; \
		rm -f $$cast; \
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
