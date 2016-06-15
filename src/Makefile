# Makefile parameters -- default values
NAME = base
TEST = simple
ARGS = -n 2

# Common paths
PATH_OBJ = .
PATH_BIN = ../bin

# Common binaries
BIN = $(PATH_BIN)/$(NAME)_$(TEST)

# Common sources ('TOOLS' in a specific order', $(wildcard tools/*.go) would not be valid)
TOOLS   = assert volatile xorshift share optik thread ttas
SOURCES = $(TOOLS:%=tools/%.go) dataset.go test/$(TEST).go
OBJECTS = $(SOURCES:%.go=$(PATH_OBJ)/%.o)

# Dataset names
DATASET = $(patsubst %.go,%,$(subst base.go,,$(wildcard *.go)))

# Compiler/linker/perf-related options
CC     = gccgo
CFLAGS = -Wall -g -O2 -I$(PATH_OBJ)
LD     = gccgo
LFLAGS = -static-libgo
PERF   = perf
PFLAGS = -e instructions:u

# Perf outputs
PERF_OUT = $(BIN)_record

.PHONY: build build-all run perf-record perf-report perf-all clean

# File rules
$(PATH_OBJ)/dataset.o: $(NAME).go
	$(CC) $(CFLAGS) -c -o $@ $<
$(PATH_OBJ)/%.o: %.go
	$(CC) $(CFLAGS) -c -o $@ $<
$(BIN): $(OBJECTS)
	$(LD) $(LFLAGS) -o $@ $^

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
	$(RM) tools/*.o test/*.o $(PATH_OBJ)/*.o $(PATH_BIN)/*
