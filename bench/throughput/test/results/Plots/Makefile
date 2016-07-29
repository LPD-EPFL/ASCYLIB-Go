PLOTS = $(subst gp,pdf,$(filter-out gp/style.gp,$(wildcard gp/*.gp)))

.PHONY: all

pdf/%.pdf: gp/%.gp
	gnuplot $^
all: $(PLOTS)
