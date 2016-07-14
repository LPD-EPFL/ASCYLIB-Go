NAME = tester
OBJ  = .
BIN  = ../bin/$(NAME)

HEADERS = $(wildcard *.hpp)
SOURCES = $(wildcard *.cpp)
OBJECTS = $(SOURCES:%.cpp=$(OBJ)/%.o)

CC     = g++
CFLAGS = -Wall -O2 -std=c++14
LFLAGS =

.PHONY: build run clean

$(OBJ)/%.o: %.cpp $(HEADERS)
	$(CC) $(CFLAGS) -c -o $@ $<
$(BIN): $(OBJECTS)
	$(CC) $(LFLAGS) -o $@ $^

build: $(BIN)
run: $(BIN)
	@$(BIN) $(ARGS)
clean:
	rm -f $(OBJECTS) $(BIN)
