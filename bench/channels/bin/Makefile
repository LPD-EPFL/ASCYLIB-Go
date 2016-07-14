.PHONY: test

# Actual test
test: tester channels
	@./tester channels

# Binary buiders
tester:
	$(MAKE) -C ../test
channels:
	$(MAKE) -C ..
