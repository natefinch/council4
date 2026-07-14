package parser

import "testing"

func TestParseGroupProtectionTypedEffects(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Until end of turn, your life total can't change, and permanents you control gain hexproof and indestructible.",
		Context{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}
	var effects []EffectSyntax
	for _, sentence := range document.Abilities[0].Sentences {
		effects = append(effects, sentence.Effects...)
	}
	if len(effects) != 2 {
		t.Fatalf("effects = %+v, want 2", effects)
	}

	life := effects[0]
	if life.Kind != EffectLifeTotalCantChange || !life.Exact ||
		life.Duration != EffectDurationUntilEndOfTurn ||
		life.Context != EffectContextController {
		t.Fatalf("effects[0] = %+v, want exact controller life-total-can't-change until end of turn", life)
	}

	group := effects[1]
	if group.Kind != EffectGain || !group.Exact ||
		group.Duration != EffectDurationUntilEndOfTurn ||
		group.Connection != EffectConnectionAnd {
		t.Fatalf("effects[1] = %+v, want exact and-connected group grant until end of turn", group)
	}
}

func TestParseGroupProtectionFailsClosedWithoutDuration(t *testing.T) {
	t.Parallel()
	// Without the until-end-of-turn scope the compound sentence is not the
	// recognized protective spell, so it must not produce the two exact effects.
	document, _ := Parse(
		"Your life total can't change, and permanents you control gain hexproof and indestructible.",
		Context{InstantOrSorcery: true},
	)
	var effects []EffectSyntax
	for _, sentence := range document.Abilities[0].Sentences {
		effects = append(effects, sentence.Effects...)
	}
	for _, effect := range effects {
		if effect.Kind == EffectLifeTotalCantChange && effect.Exact {
			t.Fatalf("effects = %+v, want no exact life-total-can't-change without duration", effects)
		}
	}
}
