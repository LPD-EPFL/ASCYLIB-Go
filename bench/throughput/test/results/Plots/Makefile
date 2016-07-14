PLOTS = $(subst gp,eps,$(filter-out gp/style.gp,$(wildcard gp/*.gp)))

.PHONY: all

eps/%.eps: gp/%.gp
	gnuplot $^
all: $(PLOTS)
