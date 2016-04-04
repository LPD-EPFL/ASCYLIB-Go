NAME = base
TEST = simple

OBJ  = .
BIN  = ../bin/$(NAME)_$(TEST)

SOURCES = $(wildcard tools/*.go) test/prototype.go dataset.go test/$(TEST).go
OBJECTS = $(SOURCES:%.go=$(OBJ)/%.o)

CC     = gccgo
CFLAGS = -Wall -g -O2 -I$(OBJ)
LFLAGS = -static-libgo

.PHONY: build run clean

$(OBJ)/dataset.o: $(NAME).go
	$(CC) $(CFLAGS) -c -o $@ $<
$(OBJ)/%.o: %.go
	$(CC) $(CFLAGS) -c -o $@ $<
$(BIN): $(OBJECTS)
	$(CC) $(LFLAGS) -o $@ $^

build: $(BIN)
run:
	@$(BIN) -n 2
clean:
	$(RM) tools/*.o test/*.o ./*.o ../bin/*
