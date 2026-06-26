package magefiles

import (
	"os"
	"testing"
)

func TestSupportedDocArgs(t *testing.T) {
	args := supportedDocArgs("/cache/oracle-cards.json", "/tmp/scratch")
	if len(args) < 2 || args[0] != "run" || args[1] != "./cardgen/oracle/cmd/compilecards" {
		t.Fatalf("args = %v, want a compilecards run invocation", args)
	}
	assertArgPair(t, args, "-in", "/cache/oracle-cards.json")
	assertArgPair(t, args, "-out", "/tmp/scratch")
	assertArgPair(t, args, "-report", os.DevNull)
	assertArgPair(t, args, "-supported", "supported.md")
}
