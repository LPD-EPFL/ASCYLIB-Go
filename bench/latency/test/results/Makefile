HOSTS  = lpd48core lpdxeon2680 lpdquad
PLOTS  = $(subst gp/,,$(subst .gp,,$(filter-out gp/style.gp,$(wildcard gp/*.gp))))
GRAPHS = $(foreach plot,$(PLOTS),$(foreach host,$(HOSTS),eps/$(plot)_$(host).eps))

.PHONY: all

all: $(GRAPHS)
clean:
	$(RM) eps/*.eps

define rule =
eps/$(1)_$(2).eps: gp/$(1).gp $(wildcard data/$(2).*)
	gnuplot -e "hostname='$(2)'" gp/$(1).gp
endef
$(foreach plot,$(PLOTS),$(foreach host,$(HOSTS),$(eval $(call rule,$(plot),$(host)))))
