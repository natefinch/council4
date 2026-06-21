package parser

import "testing"

func TestParseUnearthKeyword(t *testing.T) {
	t.Parallel()
	source := "Unearth {3}{B} ({3}{B}: Return this card from your graveyard to the " +
		"battlefield. It gains haste. Exile it at the beginning of the next end step " +
		"or if it would leave the battlefield. Unearth only as a sorcery.)"
	keywords := keywordsFor(t, source)
	if len(keywords) != 1 {
		t.Fatalf("keywords = %+v; want exactly one", keywords)
	}
	got := keywords[0]
	if got.Kind != KeywordUnearth {
		t.Fatalf("kind = %v; want %v", got.Kind, KeywordUnearth)
	}
	if got.Parameter.Kind != KeywordParameterManaCost {
		t.Fatalf("parameter kind = %v; want %v", got.Parameter.Kind, KeywordParameterManaCost)
	}
	if got.Parameter.ManaCost().String() != "{3}{B}" {
		t.Fatalf("mana cost = %q; want {3}{B}", got.Parameter.ManaCost().String())
	}
}

func TestParseUnearthKeywordVariants(t *testing.T) {
	t.Parallel()
	for cost, source := range map[string]string{
		"{7}":       "Unearth {7}",
		"{1}{R}":    "Unearth {1}{R}",
		"{5}{B}{R}": "Unearth {5}{B}{R}",
	} {
		keywords := keywordsFor(t, source)
		if len(keywords) != 1 || keywords[0].Kind != KeywordUnearth {
			t.Fatalf("%q keywords = %+v; want one Unearth", source, keywords)
		}
		if keywords[0].Parameter.ManaCost().String() != cost {
			t.Fatalf("%q mana cost = %q; want %q", source, keywords[0].Parameter.ManaCost().String(), cost)
		}
	}
}
