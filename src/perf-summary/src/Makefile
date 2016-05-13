NAME = ps
OBJ  = .
BIN  = ../bin/$(NAME)

SOURCES = ps.go
OBJECTS = $(SOURCES:%.go=$(OBJ)/%.o)

CC     = gccgo
CFLAGS = -Wall -g -O2
LD     = gccgo
LFLAGS = -static-libgo

.PHONY: build clean

%.o: %.go
	$(CC) $(CFLAGS) -c -o $@ $<
$(BIN): $(OBJECTS)
	$(LD) $(LFLAGS) -o $@ $^

build: $(BIN)
clean:
	rm -f $(OBJECTS) $(BIN)
