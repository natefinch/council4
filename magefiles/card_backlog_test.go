package magefiles

import "testing"

func TestCardBacklogArgs(t *testing.T) {
	t.Parallel()
	args := cardBacklogArgs("/cache/oracle-cards.json", "/tmp/compile-report.json", "/tmp/backlog-report.json")
	if len(args) < 2 || args[0] != "run" || args[1] != "./cardgen/oracle/cmd/cardbacklog" {
		t.Fatalf("args = %v, want a cardbacklog run invocation", args)
	}
	assertArgPair(t, args, "-in", "/cache/oracle-cards.json")
	assertArgPair(t, args, "-out", "card-backlog.md")
	assertArgPair(t, args, "-report", "/tmp/backlog-report.json")
	assertArgPair(t, args, "-compile-report", "/tmp/compile-report.json")
}
