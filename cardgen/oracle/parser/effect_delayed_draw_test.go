package parser

import "testing"

func TestParseArcaneDenialDelayedDraws(t *testing.T) {
	t.Parallel()
	source := "Counter target spell. Its controller may draw up to two cards at the beginning of the next turn's upkeep.\nYou draw a card at the beginning of the next turn's upkeep."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	var effects []EffectSyntax
	for _, ability := range document.Abilities {
		for _, sentence := range ability.Sentences {
			effects = append(effects, sentence.Effects...)
		}
	}
	if len(effects) != 3 {
		t.Fatalf("effects = %#v, want three", effects)
	}
	targetDraw := effects[1]
	if !targetDraw.Exact ||
		targetDraw.Context != EffectContextReferencedObjectController ||
		!targetDraw.Optional ||
		targetDraw.DelayedTiming != DelayedTimingNextUpkeep ||
		!targetDraw.Amount.RangeKnown ||
		targetDraw.Amount.Minimum != 0 ||
		targetDraw.Amount.Maximum != 2 {
		t.Fatalf("target-controller draw = %#v", targetDraw)
	}
	controllerDraw := effects[2]
	if !controllerDraw.Exact ||
		controllerDraw.Context != EffectContextController ||
		controllerDraw.Optional ||
		controllerDraw.DelayedTiming != DelayedTimingNextUpkeep ||
		!controllerDraw.Amount.Known ||
		controllerDraw.Amount.Value != 1 {
		t.Fatalf("controller draw = %#v", controllerDraw)
	}
}

func TestParseDelayedDrawTimingNearMissesFailClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Draw up to two cards at the beginning of the next end step.",
		"Draw up to two cards at the beginning of your next upkeep.",
		"Draw up to two cards at the beginning of each of the next two turns' upkeeps.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(source, Context{InstantOrSorcery: true})
			effect := document.Abilities[0].Sentences[0].Effects[0]
			if effect.DelayedTiming == DelayedTimingNextUpkeep && effect.Exact {
				t.Fatalf("near miss parsed as exact next-turn-upkeep draw: %#v", effect)
			}
		})
	}
}
