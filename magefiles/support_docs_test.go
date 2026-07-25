package magefiles

import (
	"os"
	"testing"
)

func TestSupportDocsArgs(t *testing.T) {
	args := supportDocsArgs("/cache/oracle-cards.json", "/tmp/scratch")
	if len(args) < 2 || args[0] != "run" || args[1] != "./cardgen/oracle/cmd/compilecards" {
		t.Fatalf("args = %v, want a compilecards run invocation", args)
	}
	assertArgPair(t, args, "-in", "/cache/oracle-cards.json")
	assertArgPair(t, args, "-out", "/tmp/scratch")
	assertArgPair(t, args, "-report", os.DevNull)
	assertDocumentationArgs(t, args)
}

// TestSupportDocsArgsMatchCardSupportDocumentation pins the invariant that broke
// the documentation set: CI regenerates docs through SupportDocs while a local
// mage cardSupport run regenerates them through CardSupport. When those two flag
// lists drifted, only supported.md stayed current and the unblock roadmap in
// unsupported-reasons.md silently went stale for weeks.
func TestSupportDocsArgsMatchCardSupportDocumentation(t *testing.T) {
	docs := supportDocsArgs("/cache/oracle-cards.json", "/tmp/scratch")
	full := cardSupportArgs("./cardgen/oracle/cmd/compilecards", "/cache/oracle-cards.json", "/tmp/generated", true)
	assertDocumentationArgs(t, docs)
	assertDocumentationArgs(t, full)
}

// TestCardSupportArgsOmitDocumentationForExternalOutput keeps CardTree from
// rewriting this repository's documentation while generating a card tree
// elsewhere.
func TestCardSupportArgsOmitDocumentationForExternalOutput(t *testing.T) {
	args := cardSupportArgs("./cardgen/oracle/cmd/compilecards", "/cache/oracle-cards.json", "/tmp/generated", false)
	for _, flag := range []string{"-supported", "-unsupported", "-unsupported-reasons", "-readme"} {
		for _, arg := range args {
			if arg == flag {
				t.Fatalf("args = %v, want no %s when documentation is disabled", args, flag)
			}
		}
	}
}

// assertDocumentationArgs asserts that args writes every committed card-support
// document.
func assertDocumentationArgs(t *testing.T, args []string) {
	t.Helper()
	assertArgPair(t, args, "-supported", "supported.md")
	assertArgPair(t, args, "-unsupported", "unsupported.md")
	assertArgPair(t, args, "-unsupported-reasons", "unsupported-reasons.md")
	assertArgPair(t, args, "-readme", "README.md")
}
