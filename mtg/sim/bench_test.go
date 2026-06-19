package sim

import "testing"

// These benchmarks track per-game runtime and batch throughput so a regression
// that would push 1,000 games past the 30-minute budget (ADR 0003) shows up as
// a measurable slowdown. They use the lightweight land-only smoke decks so a run
// finishes quickly and deterministically; a heavier board-building workload that
// exercises the uncached continuous-effect recompute (mtg/rules/continuous.go)
// is tracked separately. Record fresh numbers with:
//
//	go test ./mtg/sim/ -run '^$' -bench . -benchmem
//
// See README.md for a recorded baseline.

// BenchmarkRunOneGame measures the cost of a single game end to end, including
// allocations.
func BenchmarkRunOneGame(b *testing.B) {
	cfg := smokeConfig(1, 20240101)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = RunOne(cfg, 0)
	}
}

// BenchmarkRunBatchSequential measures a small batch driven by a single worker,
// isolating per-game cost from concurrency effects.
func BenchmarkRunBatchSequential(b *testing.B) {
	cfg := smokeConfig(16, 20240101)
	cfg.Workers = 1
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = Run(cfg)
	}
}

// BenchmarkRunBatchParallel measures the same batch across GOMAXPROCS workers,
// reporting the throughput gain from parallel execution.
func BenchmarkRunBatchParallel(b *testing.B) {
	cfg := smokeConfig(16, 20240101)
	cfg.Workers = 0 // GOMAXPROCS
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = Run(cfg)
	}
}
