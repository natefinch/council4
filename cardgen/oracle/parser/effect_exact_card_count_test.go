package parser

import "testing"

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

func TestExactCardCountFailsClosed(t *testing.T) {
	t.Parallel()
	// Each of these carries a recipient or qualifier the canonical
	// draw/discard/mill phrasing cannot faithfully reconstruct, so the
	// round-trip must fail closed rather than lower to a wrong recipient.
	cases := []struct {
		source string
		kind   EffectKind
	}{
		{"Each player discards a card at random.", EffectDiscard},
		{"Target player discards a card at random.", EffectDiscard},
		{"Target opponent discards a creature card.", EffectDiscard},
	}
	for _, c := range cases {
		if cardCountEffectExact(t, c.source, c.kind) {
			t.Errorf("cardCountEffectExact(%q) = true, want false", c.source)
		}
	}
}
