package compiler

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

// TestCompileDevourKeyword verifies that the compiler maps the parser's Devour
// as-enters replacement through to a typed EffectDevour with its multiplier
// preserved, without inspecting the printed keyword text (CR 702.81).
func TestCompileDevourKeyword(t *testing.T) {
	t.Parallel()
	source := "Devour 2 (As this creature enters, you may sacrifice any number of creatures. It enters with twice that many +1/+1 counters on it.)"
	compilation, diagnostics := compileSource(source, pipelineContext{CardName: "Devourer"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Kind != AbilityReplacement {
		t.Fatalf("kind = %v, want AbilityReplacement", ability.Kind)
	}
	if len(ability.Content.Effects) != 1 {
		t.Fatalf("effects = %#v", ability.Content.Effects)
	}
	effect := ability.Content.Effects[0]
	if effect.Kind != EffectDevour {
		t.Fatalf("effect kind = %v, want EffectDevour", effect.Kind)
	}
	if !effect.EntersDevour {
		t.Fatal("EntersDevour = false, want true")
	}
	if effect.EntersDevourMultiplier != 2 {
		t.Fatalf("EntersDevourMultiplier = %d, want 2", effect.EntersDevourMultiplier)
	}
}

// TestCompileDevourTypedKeyword verifies that the compiler carries the typed
// Devour variants' structured sacrifice filter through to the compiled effect
// without inspecting the printed keyword text (CR 702.81).
func TestCompileDevourTypedKeyword(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		source   string
		wantType types.Card
		wantSub  types.Sub
	}{
		{"artifact", "Devour artifact 1 (As this creature enters, you may sacrifice any number of artifacts. It enters with that many +1/+1 counters on it.)", types.Artifact, ""},
		{"land", "Devour land 3 (As this creature enters, you may sacrifice any number of lands. It enters with three times that many +1/+1 counters on it.)", types.Land, ""},
		{"Food", "Devour Food 3 (As this creature enters, you may sacrifice any number of Foods. It enters with three times that many +1/+1 counters on it.)", "", types.Food},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(test.source, pipelineContext{CardName: "Devourer"})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effect := compilation.Abilities[0].Content.Effects[0]
			if effect.Kind != EffectDevour {
				t.Fatalf("effect kind = %v, want EffectDevour", effect.Kind)
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
