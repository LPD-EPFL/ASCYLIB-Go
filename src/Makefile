# Makefile parameters -- default values
NAME = base
TEST = simple
ARGS = -n 2

# Common paths
PATH_OBJ = .
PATH_BIN = ../bin
PATH_PSM = perf-summary/src

# Common binaries
BIN = $(PATH_BIN)/$(NAME)_$(TEST)
PSM = perf-summary/bin/ps

# Common sources
SOURCES = tools/volatile.go tools/xorshift.go tools/assert.go tools/optik.go tools/thread.go tools/share.go dataset.go test/$(TEST).go
OBJECTS = $(SOURCES:%.go=$(PATH_OBJ)/%.o)

# Dataset names
DATASET = $(patsubst %.go,%,$(subst base.go,,$(wildcard *.go)))

# Compiler/linker/perf-related options
CC     = gccgo
CFLAGS = -Wall -g -O2 -I$(PATH_OBJ)
LD     = gccgo
LFLAGS = -static-libgo
PERF   = perf
PFLAGS =

# Perf outputs
PERF_OUT   = $(BIN)_record
PERF_DUMP  = $(PERF_OUT)_dump
PERF_TRAIL = _record_dump
PERF_SUMTR = _summary

.PHONY: build build-all clean run perf-record perf-report perf-all perf-summary

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
	@$(PERF) report -i $(PERF_OUT) $(PFLAGS)
perf-report-dump: $(PERF_OUT)
	@$(PERF) report -i $(PERF_OUT) $(PFLAGS) > $(PERF_DUMP)
perf-summary:
	@$(MAKE) -C $(PATH_PSM) build
	@$(foreach dump,$(wildcard $(PATH_BIN)/*$(PERF_TRAIL)),$(PSM) < $(dump) > $(patsubst %$(PERF_TRAIL),%$(PERF_SUMTR),$(dump));)
perf-all:
	@$(foreach name,$(DATASET),touch $(name).go; $(MAKE) perf-report-dump NAME=$(name) TEST=$(TEST) ARGS="$(ARGS)";)
	@$(MAKE) perf-summary

clean:
	$(RM) tools/*.o test/*.o $(PATH_OBJ)/*.o $(PATH_BIN)/*
	@$(MAKE) -C $(PATH_PSM) clean
