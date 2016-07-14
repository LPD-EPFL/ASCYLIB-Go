ARGS =
BIN  = bin/channels

.PHONY: build run clean

build:
	export GOPATH="$(abspath .)"; go build -o $(BIN) channels
run: $(BIN)
	@$(BIN) $(ARGS)
clean:
	go clean
