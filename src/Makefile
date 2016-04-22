NAME = base
TEST = simple

OBJ  = .
BIN  = ../bin/$(NAME)_$(TEST)

SOURCES = tools/volatile.go tools/xorshift.go tools/assert.go tools/optik.go tools/thread.go tools/share.go dataset.go test/$(TEST).go
OBJECTS = $(SOURCES:%.go=$(OBJ)/%.o)

CC     = gccgo
CFLAGS = -Wall -g -O2 -I$(OBJ)
LFLAGS = -static-libgo

.PHONY: build build-all run clean

$(OBJ)/dataset.o: $(NAME).go
	$(CC) $(CFLAGS) -c -o $@ $<
$(OBJ)/%.o: %.go
	$(CC) $(CFLAGS) -c -o $@ $<
$(BIN): $(OBJECTS)
	$(CC) $(LFLAGS) -o $@ $^

build: $(BIN)
build-all:
	@$(foreach name,$(patsubst %.go,%,$(subst base.go,,$(wildcard *.go))),touch $(name).go; $(MAKE) build NAME=$(name) TEST=$(TEST);)
run:
	@$(BIN)
clean:
	$(RM) tools/*.o test/*.o ./*.o ../bin/*
