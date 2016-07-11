HOSTS  = lpd48core lpdxeon2680 lpdquad
PLOTS  = $(subst gp/,,$(subst .gp,,$(filter-out gp/style.gp,$(wildcard gp/*.gp))))
GRAPHS = $(foreach plot,$(PLOTS),$(foreach host,$(HOSTS),pdf/$(plot)_$(host).pdf))

.PHONY: all

all: $(GRAPHS)
clean:
	$(RM) pdf/*.pdf

define rule =
pdf/$(1)_$(2).pdf: gp/$(1).gp
	gnuplot -e "hostname='$(2)'" gp/$(1).gp
endef
$(foreach plot,$(PLOTS),$(foreach host,$(HOSTS),$(eval $(call rule,$(plot),$(host)))))
