# File names
GROUP_HT = $(sort $(wildcard hashtable_*))
GROUP_LL = $(sort $(wildcard linkedlist_*))
GROUP_SL = $(sort $(wildcard skiplist_*))
GROUP_ST = $(sort $(wildcard stack_*))
GROUP_QU = $(sort $(wildcard queue_*)) $(sort $(wildcard priorityqueue_*))

.PHONY: all ht ll sl st qu

all: ht ll sl st qu

ht: tester
	@./tester ht $(GROUP_HT)
ll: tester
	@./tester ll $(GROUP_LL)
sl: tester
	@./tester sl $(GROUP_SL)
st: tester
	@./tester st $(GROUP_ST)
qu: tester
	@./tester qu $(GROUP_QU)

# Binary buiders
tester:
	$(MAKE) -C ../test
