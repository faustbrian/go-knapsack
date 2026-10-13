GO ?= go
BENCH_TIME ?= 100ms

.PHONY: benchmark docs package-test

package-test:
	$(GO) test scripts/coverage-exception/check.go scripts/coverage-exception/check_test.go
	./scripts/check-coverage.sh
	$(GO) test -race . ./solver \
		-run '^(TestProductionContainsNoGoroutineLaunches|TestSolversStopAfterRepeatedCancellation)$$' \
		-count=5
	./scripts/verify-corpus.sh
	./scripts/verify-boxpacker.sh
	./scripts/test-dependency-review.sh
	./scripts/check-dependency-review.sh --publish

docs:
	./scripts/check-docs.sh

benchmark:
	./scripts/test-benchmark-rss.sh
	$(GO) test ./... -run '^$$' -bench Benchmark -benchmem \
		-benchtime="$(BENCH_TIME)"
	./scripts/benchmark-rss.sh "$(BENCH_TIME)"
	./scripts/benchmark-compare.sh "$(BENCH_TIME)"
	./scripts/benchmark-boxpacker.sh
