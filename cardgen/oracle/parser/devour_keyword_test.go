package parser

import "testing"

func devourEffect(t *testing.T, source string) EffectSyntax {
	t.Helper()
	document, diagnostics := Parse(source, Context{CardName: "Devourer"})
	for _, diagnostic := range diagnostics {
		t.Fatalf("Parse(%q) produced diagnostic: %s", source, diagnostic.Summary)
	}
	for i := range document.Abilities {
		ability := &document.Abilities[i]
		if ability.Kind != AbilityReplacement {
			continue
		}
		for j := range ability.Sentences {
			sentence := &ability.Sentences[j]
			for k := range sentence.Effects {
				if sentence.Effects[k].Kind == EffectDevour {
					return sentence.Effects[k]
				}
			}
		}
	}
	t.Fatalf("Parse(%q) produced no Devour replacement effect", source)
	return EffectSyntax{}
}

func TestExpandDevourKeywordMultiplier(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
		want   int
	}{
		{"devour 1", "Devour 1 (As this creature enters, you may sacrifice any number of creatures. It enters with that many +1/+1 counters on it.)", 1},
		{"devour 2", "Devour 2 (As this creature enters, you may sacrifice any number of creatures. It enters with twice that many +1/+1 counters on it.)", 2},
		{"devour 3", "Devour 3 (As this creature enters, you may sacrifice any number of creatures. It enters with three times that many +1/+1 counters on it.)", 3},
		{"bare keyword", "Devour 2", 2},
		{"after other keywords", "Flying, haste\nDevour 2", 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			effect := devourEffect(t, test.source)
			if !effect.EntersDevour {
				t.Fatal("EntersDevour = false, want true")
			}
			if !effect.Exact {
				t.Fatal("Exact = false, want true")
			}
			if effect.EntersDevourMultiplier != test.want {
				t.Fatalf("EntersDevourMultiplier = %d, want %d", effect.EntersDevourMultiplier, test.want)
			}
		})
	}
}

func TestExpandDevourKeywordLeavesOtherFormsAlone(t *testing.T) {
	t.Parallel()
	// The typed and variable Devour forms must not be rewritten to the
	// creature-sacrifice canonical text.
	cases := []string{
		"Devour artifact 1 (As this creature enters, you may sacrifice any number of artifacts. It enters with that many +1/+1 counters on it.)",
		"Devour land 3 (As this creature enters, you may sacrifice any number of lands. It enters with three times that many +1/+1 counters on it.)",
		"Devour X, where X is the number of creatures devoured this way",
	}
	for _, source := range cases {
		if got := expandDevourKeyword(source); got != source {
			t.Fatalf("expandDevourKeyword(%q) = %q, want unchanged", source, got)
		}
	}
}
