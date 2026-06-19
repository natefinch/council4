package parser

import "testing"

// digSequenceEffects parses a two-sentence impulse dig body and returns its two
// resolving effects (the EffectDig look and the EffectPut put).
func digSequenceEffects(t *testing.T, source string) []EffectSyntax {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("Parse(%q) abilities = %#v", source, document.Abilities)
	}
	var effects []EffectSyntax
	for si := range document.Abilities[0].Sentences {
		effects = append(effects, document.Abilities[0].Sentences[si].Effects...)
	}
	return effects
}

func TestExactDigSequenceAccepts(t *testing.T) {
	t.Parallel()
	cases := []struct {
		source   string
		look     int
		take     int
		source2  DigSourceKind
		singular bool
	}{
		{
			"Look at the top three cards of your library. Put one of them into your hand and the rest into your graveyard.",
			3, 1, DigSourceThem, false,
		},
		{
			"Look at the top two cards of your library. Put one of them into your hand and the other into your graveyard.",
			2, 1, DigSourceThem, true,
		},
		{
			"Look at the top three cards of your library. Put two of them into your hand and the rest into your graveyard.",
			3, 2, DigSourceThem, false,
		},
		{
			"Look at the top four cards of your library. Put one of those cards into your hand and the rest into your graveyard.",
			4, 1, DigSourceThoseCards, false,
		},
		{
			"Look at the top two cards of your library. Put one into your hand and the other into your graveyard.",
			2, 1, DigSourceNone, true,
		},
	}
	for _, c := range cases {
		effects := digSequenceEffects(t, c.source)
		if len(effects) != 2 {
			t.Fatalf("Parse(%q) effects = %d, want 2", c.source, len(effects))
		}
		look, put := effects[0], effects[1]
		if look.Kind != EffectDig || !look.Exact {
			t.Errorf("Parse(%q) look effect = {kind:%s exact:%v}, want EffectDig exact", c.source, look.Kind, look.Exact)
		}
		if look.Amount.Value != c.look {
			t.Errorf("Parse(%q) look amount = %d, want %d", c.source, look.Amount.Value, c.look)
		}
		if put.Kind != EffectPut || !put.Exact || !put.Dig.Put {
			t.Errorf("Parse(%q) put effect = {kind:%s exact:%v dig:%+v}, want exact dig put", c.source, put.Kind, put.Exact, put.Dig)
		}
		if put.Amount.Value != c.take {
			t.Errorf("Parse(%q) take amount = %d, want %d", c.source, put.Amount.Value, c.take)
		}
		if put.Dig.Source != c.source2 || put.Dig.Singular != c.singular {
			t.Errorf("Parse(%q) dig = %+v, want source %q singular %v", c.source, put.Dig, c.source2, c.singular)
		}
	}
}

func TestExactDigSequenceFailsClosed(t *testing.T) {
	t.Parallel()
	// Each put clause carries a remainder destination or count the Dig primitive
	// cannot faithfully model, so its round-trip must fail closed.
	cases := []string{
		// Library-bottom remainder with an ordering rider the engine does not model.
		"Look at the top four cards of your library. Put one of them into your hand and the rest on the bottom of your library in any order.",
		"Look at the top four cards of your library. Put one of them into your hand and the rest on the bottom of your library in a random order.",
		"Look at the top two cards of your library. Put one of them into your hand and the other on the bottom of your library.",
		// "Up to" makes the take count variable.
		"Look at the top five cards of your library. Put up to two of them into your hand and the rest on the bottom of your library in a random order.",
	}
	for _, source := range cases {
		effects := digSequenceEffects(t, source)
		for i := range effects {
			if effects[i].Kind == EffectPut && effects[i].Exact && effects[i].Dig.Put {
				t.Errorf("Parse(%q) put effect lowered exact dig, want fail closed", source)
			}
		}
	}
}
