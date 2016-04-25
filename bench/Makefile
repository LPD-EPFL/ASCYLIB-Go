CC     = gccgo
CFLAGS = -Wall -g -O2
LD     = gccgo
LFLAGS = -static-libgo

.PHONY: clean

# Benchmarks (auto-generated)
define new_rule =
$(1).o: $(1).go
	$(CC) $(CFLAGS) -c -o $(1).o $(1).go
$(1): $(1).o
	$(LD) $(LFLAGS) -o $(1) $(1).o
endef
$(foreach source,$(wildcard *.go),$(eval $(call new_rule,$(patsubst %.go,%,$(source)))))

clean:
	$(RM) $(foreach name,$(patsubst %.go,%,$(wildcard *.go)),$(name) $(name).o)
