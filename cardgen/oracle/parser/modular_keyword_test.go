package parser

import (
	"strings"
	"testing"
)

func TestExpandModularKeyword(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		source   string
		counters string
	}{
		{"rank one with reminder", "Modular 1 (This creature enters with a +1/+1 counter on it. When it dies, you may put its +1/+1 counters on target artifact creature.)", "a +1/+1 counter"},
		{"rank three", "Modular 3", "three +1/+1 counters"},
		{"after another keyword", "Flying\nModular 4", "four +1/+1 counters"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := expandModularKeyword(test.source)
			wantStatic := "This creature enters with " + test.counters + " on it."
			if !containsLine(got, wantStatic) {
				t.Fatalf("expandModularKeyword(%q) = %q, want a line %q", test.source, got, wantStatic)
			}
			if !strings.Contains(got, "When this creature dies, you may move all counters from this creature onto target artifact creature.") {
				t.Fatalf("expandModularKeyword(%q) = %q, want the dies-trigger ability", test.source, got)
			}
		})
	}
}

func TestExpandModularKeywordLeavesOtherTextAlone(t *testing.T) {
	t.Parallel()
	// A variable rank ("Modular—Sunburst") and lines that only mention the word
	// must not be rewritten.
	for _, source := range []string{
		"Modular\u2014Sunburst",
		"Poison Modular 2",
		"Whenever a Modular creature dies, draw a card.",
	} {
		if got := expandModularKeyword(source); got != source {
			t.Fatalf("expandModularKeyword(%q) rewrote unrelated text: %q", source, got)
		}
	}
}

func TestModularExpandsToConjunctiveArtifactCreatureTarget(t *testing.T) {
	t.Parallel()
	doc, diags := Parse("Modular 2", Context{CardName: "Test Modular"})
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	var found bool
	for _, ability := range doc.Abilities {
		for s := range ability.Sentences {
			for e := range ability.Sentences[s].Effects {
				for _, target := range ability.Sentences[s].Effects[e].Targets {
					if !target.Selection.ConjunctiveTypes {
						continue
					}
					if !target.Exact {
						t.Fatalf("conjunctive artifact-creature target is not exact: %q", target.Text)
					}
					found = true
				}
			}
		}
	}
	if !found {
		t.Fatal("Modular expansion produced no conjunctive artifact-creature target")
	}
}
