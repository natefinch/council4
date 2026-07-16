package parser

import "testing"

// TestParseOnlyYouDrawIsExactControllerDraw proves the emphatic "only you draw"
// controller subject round-trips as an exact card draw, the voice Kumena's
// Awakening uses for its conditional replacement ("... instead only you draw a
// card."). The leading "only" is an emphatic adverb on the controller subject
// and adds no rules meaning beyond an ordinary controller draw.
func TestParseOnlyYouDrawIsExactControllerDraw(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse("Only you draw a card.", Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 {
		t.Fatalf("effects = %d, want 1", len(effects))
	}
	if effects[0].Kind != EffectDraw {
		t.Fatalf("effect kind = %v, want EffectDraw", effects[0].Kind)
	}
	if effects[0].Context != EffectContextController {
		t.Fatalf("effect context = %v, want EffectContextController", effects[0].Context)
	}
	if !effects[0].Exact {
		t.Fatal("only-you controller draw not marked exact")
	}
}

// TestParseKumenasAwakeningUpkeepBody proves the full upkeep body parses into two
// exact draw effects — an each-player draw followed by a controller "instead"
// replacement — gated by a city's-blessing condition.
func TestParseKumenasAwakeningUpkeepBody(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"At the beginning of your upkeep, each player draws a card. "+
			"If you have the city's blessing, instead only you draw a card.",
		Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %d, want 1", len(document.Abilities))
	}
	ability := document.Abilities[0]
	if ability.Trigger == nil {
		t.Fatal("upkeep body did not parse as a triggered ability")
	}

	var effects []EffectSyntax
	for _, sentence := range ability.Sentences {
		effects = append(effects, sentence.Effects...)
	}
	if len(effects) != 2 {
		t.Fatalf("effects = %d, want 2: %#v", len(effects), effects)
	}

	base := effects[0]
	if base.Kind != EffectDraw || base.Context != EffectContextEachPlayer || !base.Exact {
		t.Fatalf("base effect = %+v, want exact each-player draw", base)
	}
	if base.Replacement.Kind != EffectReplacementNone {
		t.Fatalf("base replacement = %v, want none", base.Replacement.Kind)
	}

	replacement := effects[1]
	if replacement.Kind != EffectDraw || replacement.Context != EffectContextController || !replacement.Exact {
		t.Fatalf("replacement effect = %+v, want exact controller draw", replacement)
	}
	if replacement.Replacement.Kind != EffectReplacementInstead {
		t.Fatalf("replacement kind = %v, want instead", replacement.Replacement.Kind)
	}

	found := false
	for _, clause := range ability.ConditionClauses {
		if clause.Predicate == ConditionPredicateControllerHasCityBlessing {
			found = true
		}
	}
	if !found {
		t.Fatalf("city's-blessing condition not recognized: %#v", ability.ConditionClauses)
	}
}
