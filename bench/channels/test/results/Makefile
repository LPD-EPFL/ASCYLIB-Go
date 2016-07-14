INPUTS = $(sort $(subst .dat,,$(subst data/,,$(wildcard data/*.dat))))

.PHONY: default all

default: all

# Create rules for a single input file
define rule_single =
GRAPHS += eps/single.$(1).$(2).s$(3).eps
eps/single.$(1).$(2).s$(3).eps: gp/single.gp data/$(1).$(2).s$(3).dat
	gnuplot -e "hostname='$(1)'" -e "mode='$(2)'" -e "servers=$(3)" gp/single.gp
endef
$(foreach input,$(INPUTS),$(eval $(call rule_single,$(word 1,$(subst ., ,$(input))),$(word 2,$(subst ., ,$(input))),$(subst s,,$(word 3,$(subst ., ,$(input)))))))

all: $(GRAPHS)
clean:
	$(RM) eps/*.eps
