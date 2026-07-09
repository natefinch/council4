package parser

import "testing"

// findExcessRedirectEffect returns the deal-damage effect a card's abilities
// carry for an "excess damage is dealt to ... instead" redirect (the one whose
// amount is the excess damage dealt this way), or nil when none parsed.
func findExcessRedirectEffect(document Document) *EffectSyntax {
	for ai := range document.Abilities {
		for si := range document.Abilities[ai].Sentences {
			sentence := &document.Abilities[ai].Sentences[si]
			for ei := range sentence.Effects {
				effect := &sentence.Effects[ei]
				if effect.Kind == EffectDealDamage &&
					effect.Amount.DynamicKind == EffectDynamicAmountExcessDamageDealtThisWay {
					return effect
				}
			}
		}
	}
	return nil
}

// TestParseSourceTrampleExcessRiderExactness proves Ram Through's conditional
// "If the creature you control has trample, excess damage is dealt to that
// creature's controller instead." rider parses to an exact excess-damage
// redirect that carries the RequireSourceTrample marker and targets the prior
// creature's controller, while Flame Spill's unconditional redirect parses to
// the same redirect without the marker (regression guard).
func TestParseSourceTrampleExcessRiderExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		source         string
		requireTrample bool
	}{
		{
			"Ram Through",
			"Target creature you control deals damage equal to its power to target creature you don't control. " +
				"If the creature you control has trample, excess damage is dealt to that creature's controller instead.",
			true,
		},
		{
			"Flame Spill",
			"Flame Spill deals 4 damage to target creature. Excess damage is dealt to that creature's controller instead.",
			false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			redirect := findExcessRedirectEffect(document)
			if redirect == nil {
				t.Fatalf("no excess-damage redirect parsed from %q", test.source)
			}
			if !redirect.Exact {
				t.Fatal("redirect Exact = false, want true")
			}
			if redirect.RequireSourceTrample != test.requireTrample {
				t.Fatalf("RequireSourceTrample = %v, want %v", redirect.RequireSourceTrample, test.requireTrample)
			}
			if redirect.DamageRecipient.Reference != DamageRecipientReferenceController {
				t.Fatalf("recipient reference = %v, want controller", redirect.DamageRecipient.Reference)
			}
		})
	}
}

// TestParseSourceTrampleExcessRiderClearsConditionSemantics proves the strip
// pass removes the standalone "If ... has trample" condition and the bare
// "trample" keyword the independent scans would otherwise surface on Ram
// Through's ability, leaving the marked effect to own the whole sentence.
func TestParseSourceTrampleExcessRiderClearsConditionSemantics(t *testing.T) {
	t.Parallel()
	document, _ := Parse(
		"Target creature you control deals damage equal to its power to target creature you don't control. "+
			"If the creature you control has trample, excess damage is dealt to that creature's controller instead.",
		Context{InstantOrSorcery: true})
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %d, want 1", len(document.Abilities))
	}
	ability := &document.Abilities[0]
	if ability.ConditionBoundaries != nil {
		t.Fatalf("ConditionBoundaries = %#v, want nil", ability.ConditionBoundaries)
	}
	if ability.ConditionSegments != nil {
		t.Fatalf("ConditionSegments = %#v, want nil", ability.ConditionSegments)
	}
	if ability.TriggerConditionSegments != nil {
		t.Fatalf("TriggerConditionSegments = %#v, want nil", ability.TriggerConditionSegments)
	}
	if ability.ConditionClauses != nil {
		t.Fatalf("ConditionClauses = %#v, want nil", ability.ConditionClauses)
	}
	if ability.EventHistoryConditions != nil {
		t.Fatalf("EventHistoryConditions = %#v, want nil", ability.EventHistoryConditions)
	}
	if ability.SemanticKeywords != nil {
		t.Fatalf("SemanticKeywords = %#v, want nil", ability.SemanticKeywords)
	}
}
