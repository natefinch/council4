package parser

import "testing"

// TestParseExileEntireHandIsExact verifies the involuntary whole-hand exile
// clause "Exile all cards from your hand." (Wormfang Behemoth) parses as an
// exact EffectExile carrying the ExileEntireHand flag, so lowering can emit the
// linked entire-hand exile.
func TestParseExileEntireHandIsExact(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse("Exile all cards from your hand.", Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effect := document.Abilities[0].Sentences[0].Effects[0]
	if effect.Kind != EffectExile || !effect.Exact || !effect.ExileEntireHand {
		t.Fatalf("effect = %#v, want exact entire-hand exile", effect)
	}
}

// TestParseReturnExiledCardsToHandIsExact verifies the leaves-the-battlefield
// clause "Return the exiled cards to their owner's hand." (Wormfang Behemoth)
// parses as an exact EffectReturn carrying the ReturnExiledCardsToHand flag, so
// lowering can emit the linked return to hand.
func TestParseReturnExiledCardsToHandIsExact(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse("Return the exiled cards to their owner's hand.", Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effect := document.Abilities[0].Sentences[0].Effects[0]
	if effect.Kind != EffectReturn || !effect.Exact || !effect.ReturnExiledCardsToHand {
		t.Fatalf("effect = %#v, want exact return exiled cards to hand", effect)
	}
}

// TestParseSingleExiledCardReturnIsNotEntireHandReturn guards that the singular
// O-Ring battlefield return is not mistaken for the entire-hand return: it must
// not carry the ReturnExiledCardsToHand flag.
func TestParseSingleExiledCardReturnIsNotEntireHandReturn(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Return the exiled card to the battlefield under its owner's control.",
		Context{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effect := document.Abilities[0].Sentences[0].Effects[0]
	if effect.ReturnExiledCardsToHand {
		t.Fatalf("effect = %#v, want ReturnExiledCardsToHand unset for battlefield return", effect)
	}
}
