package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/counter"
)

// cardCountEffectExact parses a single draw/discard/mill sentence and reports
// whether its resolving effect round-tripped to an exact, lowerable production.
func cardCountEffectExact(t *testing.T, source string, kind EffectKind) bool {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 || effects[0].Kind != kind {
		t.Fatalf("Parse(%q) effects = %#v", source, effects)
	}
	return effects[0].Exact
}

func TestExactCardCountGroupAndTargetAccepts(t *testing.T) {
	t.Parallel()
	cases := []struct {
		source string
		kind   EffectKind
	}{
		{"Each player draws two cards.", EffectDraw},
		{"Each opponent draws a card.", EffectDraw},
		{"Each player discards a card.", EffectDiscard},
		{"Each opponent discards two cards.", EffectDiscard},
		{"Each player mills three cards.", EffectMill},
		{"Each opponent mills two cards.", EffectMill},
		{"Target opponent discards a card.", EffectDiscard},
		{"Target opponent discards two cards.", EffectDiscard},
		{"Target player discards a card.", EffectDiscard},
		{"Target opponent mills two cards.", EffectMill},
		{"Target player draws a card.", EffectDraw},
	}
	for _, c := range cases {
		if !cardCountEffectExact(t, c.source, c.kind) {
			t.Errorf("cardCountEffectExact(%q) = false, want true", c.source)
		}
	}
}

func TestExactCardCountControllerAndTargetAccepts(t *testing.T) {
	t.Parallel()
	// "You and target <player> each draw N cards": the controller and a single
	// player target both draw. The split-on-"and" subject still classifies as
	// the controller-and-target recipient and round-trips exactly.
	cases := []struct {
		source  string
		context EffectContextKind
	}{
		{"You and target opponent each draw a card.", EffectContextControllerAndTarget},
		{"You and target opponent each draw three cards.", EffectContextControllerAndTarget},
		{"You and target player each draw two cards.", EffectContextControllerAndTarget},
	}
	for _, c := range cases {
		if !cardCountEffectExact(t, c.source, EffectDraw) {
			t.Errorf("cardCountEffectExact(%q) = false, want true", c.source)
		}
		document, _ := Parse(c.source, Context{InstantOrSorcery: true})
		got := document.Abilities[0].Sentences[0].Effects[0].Context
		if got != c.context {
			t.Errorf("Parse(%q) context = %v, want %v", c.source, got, c.context)
		}
	}
}

func TestExactCardCountControllerAndTargetFailsClosed(t *testing.T) {
	t.Parallel()
	// A non-player target ("each creature") is not a recipient the draw
	// recipient set can faithfully reconstruct, so it must not round-trip as the
	// controller-and-target form.
	const source = "You and target opponent each draw a creature card."
	if cardCountEffectExact(t, source, EffectDraw) {
		t.Errorf("cardCountEffectExact(%q) = true, want false", source)
	}
}

func TestExactCardCountControllerAndReferencedPlayerAccepts(t *testing.T) {
	t.Parallel()
	// "You and that player each draw ...": the controller and a player named by
	// a trigger reference ("that player") both draw. The split-on-"and" subject
	// classifies as the controller-and-referenced-player recipient.
	cases := []string{
		"You and that player each draw two cards.",
		"You and that player each draw that many cards.",
	}
	for _, source := range cases {
		document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
		if len(diagnostics) != 0 {
			t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
		}
		if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
			t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
		}
		effects := document.Abilities[0].Sentences[0].Effects
		if len(effects) != 1 || effects[0].Kind != EffectDraw {
			t.Fatalf("Parse(%q) effects = %#v", source, effects)
		}
		if got := effects[0].Context; got != EffectContextControllerAndReferencedPlayer {
			t.Errorf("Parse(%q) context = %v, want %v", source, got, EffectContextControllerAndReferencedPlayer)
		}
	}
}

func TestExactCardCountFailsClosed(t *testing.T) {
	t.Parallel()
	// Each of these carries a recipient or qualifier the canonical
	// draw/discard/mill phrasing cannot faithfully reconstruct, so the
	// round-trip must fail closed rather than lower to a wrong recipient.
	cases := []struct {
		source string
		kind   EffectKind
	}{
		{"Target opponent discards a creature card.", EffectDiscard},
	}
	for _, c := range cases {
		if cardCountEffectExact(t, c.source, c.kind) {
			t.Errorf("cardCountEffectExact(%q) = true, want false", c.source)
		}
	}
}

func TestExactCardCountCounterQualifiedAccepts(t *testing.T) {
	t.Parallel()
	const source = "Draw a card for each creature you control with a +1/+1 counter on it."
	if !cardCountEffectExact(t, source, EffectDraw) {
		t.Fatalf("cardCountEffectExact(%q) = false, want true", source)
	}
	document, _ := Parse(source, Context{InstantOrSorcery: true})
	effect := document.Abilities[0].Sentences[0].Effects[0]
	if effect.Amount.Selection == nil {
		t.Fatal("count amount carries no selection")
	}
	selection := effect.Amount.Selection
	if !selection.CounterRequired || selection.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("selection counter = (%v,%v), want required +1/+1", selection.CounterRequired, selection.CounterKind)
	}
}
