# Makefile parameters -- default values
NAME = base
TEST = simple
ARGS = -n 2

# Common paths
PATH_OBJ = .
PATH_BIN = ../bin

# Common binaries
BIN = $(PATH_BIN)/$(NAME)_$(TEST)

# Dataset names
DATASET = $(patsubst %.go,%,$(subst base.go,,$(wildcard *.go)))

# Compiler/linker/perf-related options
CC     = go build
CFLAGS =
PERF   = perf
PFLAGS = -e instructions:u
LN     = ln -s

# Perf outputs
PERF_OUT = $(BIN)_record

.PHONY: build build-all run perf-record perf-report perf-all clean

# File rules
$(BIN): Makefile $(NAME).go $(wildcard test/$(TEST)/*) $(wildcard tools/*/*.go)
	$(RM) dataset/dataset.go; $(LN) ../$(NAME).go dataset/dataset.go
	export GOPATH="$(abspath ..)"; $(CC) $(CFLAGS) -o $(BIN) test/$(TEST)

# Command rules
build: $(BIN)
build-all:
	@$(foreach name,$(DATASET),touch $(name).go; $(MAKE) build NAME=$(name) TEST=$(TEST);)

run: $(BIN)
	@$(BIN) $(ARGS)

perf-record: $(BIN)
	@$(PERF) record -o $(PERF_OUT) $(PFLAGS) -- $(BIN) $(ARGS)
$(PERF_OUT): perf-record
perf-report: $(PERF_OUT)
	@$(PERF) report -i $(PERF_OUT)
perf-all:
	@$(foreach name,$(DATASET),touch $(name).go; $(MAKE) perf-record NAME=$(name) TEST=$(TEST) ARGS="$(ARGS)";)

clean:
	$(RM) $(PATH_BIN)/*
