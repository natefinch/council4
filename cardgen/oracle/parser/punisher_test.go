package parser

import (
	"testing"
)

func punisherEffect(t *testing.T, source, cardName string) EffectSyntax {
	t.Helper()
	document, diagnostics := Parse(source, Context{CardName: cardName})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	var effects []EffectSyntax
	for i := range document.Abilities {
		for _, sentence := range document.Abilities[i].Sentences {
			effects = append(effects, sentence.Effects...)
		}
	}
	if len(effects) != 1 {
		t.Fatalf("effects = %#v, want exactly one", effects)
	}
	return effects[0]
}

func TestParsePunisherSacrificeOrDiscard(t *testing.T) {
	t.Parallel()
	effect := punisherEffect(t,
		"At the beginning of your upkeep, each opponent loses 3 life unless that player sacrifices a nonland permanent or discards a card.",
		"Hag of Ceaseless Torment")
	if effect.Kind != EffectPunisherLoseLife {
		t.Fatalf("kind = %v, want EffectPunisherLoseLife", effect.Kind)
	}
	if effect.Context != EffectContextEachOpponent {
		t.Fatalf("context = %v, want EffectContextEachOpponent", effect.Context)
	}
	if effect.Amount.Value != 3 || !effect.Amount.Known {
		t.Fatalf("amount = %+v, want known 3", effect.Amount)
	}
	if !effect.PunisherSacrifice || !effect.PunisherDiscard {
		t.Fatalf("flags sacrifice=%v discard=%v, want both true", effect.PunisherSacrifice, effect.PunisherDiscard)
	}
	excluded := effect.Selection.ExcludedTypes
	if len(excluded) != 1 || excluded[0] != CardTypeLand {
		t.Fatalf("sacrifice selection excluded types = %v, want [Land]", excluded)
	}
}

func TestParsePunisherDiscardOnly(t *testing.T) {
	t.Parallel()
	effect := punisherEffect(t,
		"Each opponent loses 2 life unless that player discards a card.",
		"Test Punisher")
	if effect.Kind != EffectPunisherLoseLife {
		t.Fatalf("kind = %v, want EffectPunisherLoseLife", effect.Kind)
	}
	if effect.PunisherSacrifice {
		t.Fatal("PunisherSacrifice = true, want false for discard-only")
	}
	if !effect.PunisherDiscard {
		t.Fatal("PunisherDiscard = false, want true")
	}
}
