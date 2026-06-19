package parser

import "testing"

// sourceSpellReductionEffect returns the first effect across an ability's
// sentences, used by the source-spell cost-reduction exactness tests.
func sourceSpellReductionEffect(t *testing.T, source string, context Context) *EffectSyntax {
	t.Helper()
	document, _ := Parse(source, context)
	if len(document.Abilities) == 0 {
		t.Fatalf("no abilities parsed from %q", source)
	}
	for ai := range document.Abilities {
		ability := &document.Abilities[ai]
		for si := range ability.Sentences {
			sentence := &ability.Sentences[si]
			for ei := range sentence.Effects {
				if sentence.Effects[ei].SourceSpellCostReduction {
					return &sentence.Effects[ei]
				}
			}
		}
	}
	return nil
}

func TestParseSourceSpellCostReductionExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		source  string
		context Context
		amount  int
	}{
		{
			name:    "each creature on the battlefield",
			source:  "This spell costs {1} less to cast for each creature on the battlefield.",
			context: Context{InstantOrSorcery: true},
			amount:  1,
		},
		{
			name:    "each creature you control",
			source:  "This spell costs {2} less to cast for each creature you control.",
			context: Context{InstantOrSorcery: true},
			amount:  2,
		},
		{
			name:    "each creature your opponents control",
			source:  "This spell costs {1} less to cast for each creature your opponents control.",
			context: Context{InstantOrSorcery: true},
			amount:  1,
		},
		{
			name:    "each artifact you control",
			source:  "This spell costs {3} less to cast for each artifact you control.",
			context: Context{InstantOrSorcery: true},
			amount:  3,
		},
		{
			name:    "self name subject",
			source:  "Draco costs {1} less to cast for each creature on the battlefield.",
			context: Context{InstantOrSorcery: true, CardName: "Draco"},
			amount:  1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			effect := sourceSpellReductionEffect(t, test.source, test.context)
			if effect == nil {
				t.Fatalf("source %q did not yield a source-spell cost reduction", test.source)
			}
			if effect.SourceSpellCostReductionAmount != test.amount {
				t.Fatalf("reduction amount = %d, want %d", effect.SourceSpellCostReductionAmount, test.amount)
			}
			if effect.Amount.Selection == nil {
				t.Fatalf("count selection missing for %q", test.source)
			}
		})
	}
}

func TestParseSourceSpellCostReductionFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		source  string
		context Context
	}{
		{
			name:    "graveyard count",
			source:  "This spell costs {1} less to cast for each creature card in your graveyard.",
			context: Context{InstantOrSorcery: true},
		},
		{
			name:    "variable amount",
			source:  "This spell costs {X} less to cast for each creature on the battlefield.",
			context: Context{InstantOrSorcery: true},
		},
		{
			name:    "increase wording",
			source:  "This spell costs {1} more to cast for each creature on the battlefield.",
			context: Context{InstantOrSorcery: true},
		},
		{
			name:    "multi sentence ability",
			source:  "This spell costs {1} less to cast for each creature on the battlefield. Draw a card.",
			context: Context{InstantOrSorcery: true},
		},
		{
			name:    "opponent player count",
			source:  "This spell costs {1} less to cast for each opponent you have.",
			context: Context{InstantOrSorcery: true},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if effect := sourceSpellReductionEffect(t, test.source, test.context); effect != nil {
				t.Fatalf("source %q was recognized as a source-spell cost reduction (amount %d)", test.source, effect.SourceSpellCostReductionAmount)
			}
		})
	}
}
