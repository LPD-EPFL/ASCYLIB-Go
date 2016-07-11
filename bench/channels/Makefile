CC     = gccgo
CFLAGS = -Wall -g -O2
LD     = gccgo
LFLAGS = -static-libgo

# Tools module used
TOOLS  = volatile xorshift
REPTLS = ../src/tools
OBJTLS = $(foreach name,$(TOOLS),$(REPTLS)/$(name).o)

.PHONY: clean

# Tools module compilation rule
$(REPTLS)/%.o: $(REPTLS)/%.go
	cd $(REPTLS); $(CC) $(CFLAGS) -c -o $(@F) $(^F)

# Benchmarks (auto-generated)
define new_rule =
$(1).o: $(1).go $(OBJTLS)
	$(CC) $(CFLAGS) -c -o $(1).o $(1).go
$(1): $(1).o
	$(LD) $(LFLAGS) -o $(1) $(1).o $(OBJTLS)
endef
$(foreach source,$(wildcard *.go),$(eval $(call new_rule,$(patsubst %.go,%,$(source)))))

clean:
	$(RM) $(foreach name,$(patsubst %.go,%,$(wildcard *.go)),$(name) $(name).o) $(OBJTLS)
