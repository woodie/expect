.PHONY: build lint test check

# No install target -- expect is a library, nothing to copy to ~/go/bin.
build:
	go build ./...

# lint and test are always verbose. check is terse (silent on pass, full
# log on any failure) -- matching every other lint/test/check split in
# this account (see gorderly's/humane's own Makefiles).

lint:
	golangci-lint run

# Verbose on purpose -- gorderly's documentation-style output, the same
# convention humane's Makefile uses for its own spec+expect suite.
test:
	go test -v ./... | gorderly -fd

# Terser than `test` on purpose: plain `go test` has no per-test dot mode
# of its own -- this just suppresses output on success and dumps the full
# log on failure, guaranteeing errors are never hidden.
check: lint
	@LOG=$$(mktemp); \
	if go test ./... > "$$LOG" 2>&1; then \
		echo "PASS"; \
	else \
		cat "$$LOG"; \
		rm -f "$$LOG"; \
		exit 1; \
	fi; \
	rm -f "$$LOG"
