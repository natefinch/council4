package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

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

func TestExpandDevourKeywordTypedVariants(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		source   string
		wantN    int
		wantType types.Card
		wantSub  types.Sub
	}{
		{"devour artifact 1", "Devour artifact 1 (As this creature enters, you may sacrifice any number of artifacts. It enters with that many +1/+1 counters on it.)", 1, types.Artifact, ""},
		{"devour land 3", "Devour land 3 (As this creature enters, you may sacrifice any number of lands. It enters with three times that many +1/+1 counters on it.)", 3, types.Land, ""},
		{"devour Food 3", "Devour Food 3 (As this creature enters, you may sacrifice any number of Foods. It enters with three times that many +1/+1 counters on it.)", 3, "", types.Food},
		{"bare devour land", "Devour land 2", 2, types.Land, ""},
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
			if effect.EntersDevourMultiplier != test.wantN {
				t.Fatalf("EntersDevourMultiplier = %d, want %d", effect.EntersDevourMultiplier, test.wantN)
			}
			if effect.EntersDevourType != test.wantType {
				t.Fatalf("EntersDevourType = %q, want %q", effect.EntersDevourType, test.wantType)
			}
			if effect.EntersDevourSubtype != test.wantSub {
				t.Fatalf("EntersDevourSubtype = %q, want %q", effect.EntersDevourSubtype, test.wantSub)
			}
		})
	}
}

func TestExpandDevourKeywordLeavesOtherFormsAlone(t *testing.T) {
	t.Parallel()
	// The variable Devour form and Devour lines naming an unsupported permanent
	// type must not be rewritten to a canonical sacrifice replacement.
	cases := []string{
		"Devour X, where X is the number of creatures devoured this way",
		"Devour Treasure 1 (As this creature enters, you may sacrifice any number of Treasures. It enters with that many +1/+1 counters on it.)",
	}
	for _, source := range cases {
		if got := expandDevourKeyword(source); got != source {
			t.Fatalf("expandDevourKeyword(%q) = %q, want unchanged", source, got)
		}
	}
}
