package magefiles

import (
	"os"
	"slices"
	"testing"
)

func TestSupportDocsArgs(t *testing.T) {
	t.Parallel()
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
//
// Both paths now share documentationArgs, so this compares the documentation
// flags each path actually emits rather than only checking each list separately.
// That still fails if a future change reintroduces an inline list in one path.
func TestSupportDocsArgsMatchCardSupportDocumentation(t *testing.T) {
	t.Parallel()
	docs := documentationFlags(t, supportDocsArgs("/cache/oracle-cards.json", "/tmp/scratch"))
	full := documentationFlags(t, cardSupportArgs("./cardgen/oracle/cmd/compilecards", "/cache/oracle-cards.json", "/tmp/generated", true))
	if !slices.Equal(docs, full) {
		t.Fatalf("supportDocsArgs documentation flags = %v, cardSupportArgs documentation flags = %v, want identical", docs, full)
	}
	assertDocumentationArgs(t, docs)
}

// documentationFlags returns the documentationArgs suffix of args, failing if
// args does not end with it. Requiring a suffix rather than a subset keeps a
// caller from interleaving its own document flags with the shared set.
func documentationFlags(t *testing.T, args []string) []string {
	t.Helper()
	want := documentationArgs()
	if len(args) < len(want) {
		t.Fatalf("args = %v, want at least the %d documentation arguments", args, len(want))
	}
	got := args[len(args)-len(want):]
	if !slices.Equal(got, want) {
		t.Fatalf("args tail = %v, want documentationArgs %v", got, want)
	}
	return got
}

// TestCardSupportArgsOmitDocumentationForExternalOutput keeps CardTree from
// rewriting this repository's documentation while generating a card tree
// elsewhere.
func TestCardSupportArgsOmitDocumentationForExternalOutput(t *testing.T) {
	t.Parallel()
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
