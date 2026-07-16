package parser

import (
	"testing"
)

// TestParseAdditionalUpkeepStepEffect proves the parser recognizes the
// extra-upkeep-step insertion wording (Paradox Haze) in both the "that player"
// and controller-scoped "you" subject forms.
func TestParseAdditionalUpkeepStepEffect(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"That player gets an additional upkeep step after this step.",
		"You get an additional upkeep step after this step.",
	} {
		document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
		if len(diagnostics) != 0 {
			t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
		}
		effects := document.Abilities[0].Sentences[0].Effects
		if len(effects) != 1 {
			t.Fatalf("Parse(%q) effects = %#v, want one", source, effects)
		}
		effect := effects[0]
		if effect.Kind != EffectAdditionalUpkeepStep {
			t.Errorf("Parse(%q) kind = %v, want EffectAdditionalUpkeepStep", source, effect.Kind)
		}
		if !effect.Exact {
			t.Errorf("Parse(%q) Exact = false, want true", source)
		}
		if !effect.AdditionalUpkeepStep {
			t.Errorf("Parse(%q) AdditionalUpkeepStep = false, want true", source)
		}
	}
}

// TestParseAdditionalUpkeepStepEffectRejects proves the parser fails closed for
// near-miss wordings so the generic effect parser handles them instead.
func TestParseAdditionalUpkeepStepEffectRejects(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"That player gets an additional combat step after this step.",
		"That player gets an additional upkeep step.",
		"That player gets two additional upkeep steps after this step.",
	} {
		document, _ := Parse(source, Context{InstantOrSorcery: true})
		if len(document.Abilities) == 0 || len(document.Abilities[0].Sentences) == 0 {
			continue
		}
		effects := document.Abilities[0].Sentences[0].Effects
		for _, effect := range effects {
			if effect.Kind == EffectAdditionalUpkeepStep {
				t.Errorf("Parse(%q) unexpectedly matched EffectAdditionalUpkeepStep", source)
			}
		}
	}
}

// TestParseEnchantedPlayerFirstUpkeepTriggerClause proves the trigger clause "At
// the beginning of enchanted player's first upkeep each turn" parses to an upkeep
// beginning-of-step trigger scoped to the enchanted player and flagged as the
// first occurrence each turn.
func TestParseEnchantedPlayerFirstUpkeepTriggerClause(t *testing.T) {
	t.Parallel()
	source := "At the beginning of enchanted player's first upkeep each turn, that player gets an additional upkeep step after this step."
	document, diagnostics := Parse(source, Context{CardName: "Paradox Haze"})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %d, want 1", len(document.Abilities))
	}
	trigger := document.Abilities[0].Trigger
	if trigger == nil {
		t.Fatal("expected a trigger clause")
	}
	if trigger.PhaseStep == nil {
		t.Fatalf("expected a phase/step trigger clause, got %#v", trigger)
	}
	clause := trigger.PhaseStep
	if clause.Player.Kind != TriggerPlayerSelectorEnchantedPlayer {
		t.Errorf("player selector = %v, want TriggerPlayerSelectorEnchantedPlayer", clause.Player.Kind)
	}
	if clause.Name.Kind != PhaseStepNameUpkeep {
		t.Errorf("phase/step name = %v, want PhaseStepNameUpkeep", clause.Name.Kind)
	}
	if !clause.First {
		t.Error("First = false, want true")
	}
	if !clause.EachTurn {
		t.Error("EachTurn = false, want true")
	}
}
