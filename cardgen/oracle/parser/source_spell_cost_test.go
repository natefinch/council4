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
		{
			name:    "creature card in your graveyard",
			source:  "This spell costs {1} less to cast for each creature card in your graveyard.",
			context: Context{InstantOrSorcery: true},
			amount:  1,
		},
		{
			name:    "land card in your graveyard",
			source:  "This spell costs {1} less to cast for each land card in your graveyard.",
			context: Context{InstantOrSorcery: true},
			amount:  1,
		},
		{
			name:    "artifact card in your hand",
			source:  "This spell costs {2} less to cast for each artifact card in your hand.",
			context: Context{InstantOrSorcery: true},
			amount:  2,
		},
		{
			name:    "historic card in your graveyard",
			source:  "This spell costs {1} less to cast for each historic card in your graveyard.",
			context: Context{InstantOrSorcery: true},
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
			name:    "library count",
			source:  "This spell costs {1} less to cast for each creature card in your library.",
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

// sourceSpellReductionDynamicEffect returns the first effect across an ability's
// sentences that is marked as a dynamic source-spell cost reduction.
func sourceSpellReductionDynamicEffect(t *testing.T, source string, context Context) *EffectSyntax {
	t.Helper()
	document, _ := Parse(source, context)
	for ai := range document.Abilities {
		ability := &document.Abilities[ai]
		for si := range ability.Sentences {
			sentence := &ability.Sentences[si]
			for ei := range sentence.Effects {
				if sentence.Effects[ei].SourceSpellCostReductionDynamic {
					return &sentence.Effects[ei]
				}
			}
		}
	}
	return nil
}

func TestParseSourceSpellCostReductionDynamicExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		source  string
		context Context
		kind    EffectDynamicAmountKind
	}{
		{
			name:    "greatest power among creatures you control",
			source:  "This spell costs {X} less to cast, where X is the greatest power among creatures you control.",
			context: Context{InstantOrSorcery: true},
			kind:    EffectDynamicAmountGreatestPower,
		},
		{
			name:    "self name subject",
			source:  "Draco costs {X} less to cast, where X is the greatest power among creatures you control.",
			context: Context{InstantOrSorcery: true, CardName: "Draco"},
			kind:    EffectDynamicAmountGreatestPower,
		},
		{
			name:    "total power of creatures you control",
			source:  "This spell costs {X} less to cast, where X is the total power of creatures you control.",
			context: Context{InstantOrSorcery: true},
			kind:    EffectDynamicAmountTotalPower,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			effect := sourceSpellReductionDynamicEffect(t, test.source, test.context)
			if effect == nil {
				t.Fatalf("source %q did not yield a dynamic source-spell cost reduction", test.source)
			}
			if effect.Amount.DynamicForm != EffectDynamicAmountFormWhereX {
				t.Fatalf("dynamic form = %q, want WhereX", effect.Amount.DynamicForm)
			}
			if effect.Amount.DynamicKind != test.kind {
				t.Fatalf("dynamic kind = %q, want %q", effect.Amount.DynamicKind, test.kind)
			}
		})
	}
}

func TestParseSourceSpellCostReductionDynamicFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		source  string
		context Context
	}{
		{
			name:    "fixed numeric symbol",
			source:  "This spell costs {2} less to cast, where X is the greatest power among creatures you control.",
			context: Context{InstantOrSorcery: true},
		},
		{
			name:    "increase wording",
			source:  "This spell costs {X} more to cast, where X is the greatest power among creatures you control.",
			context: Context{InstantOrSorcery: true},
		},
		{
			name:    "multi sentence ability",
			source:  "This spell costs {X} less to cast, where X is the greatest power among creatures you control. Draw a card.",
			context: Context{InstantOrSorcery: true},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if effect := sourceSpellReductionDynamicEffect(t, test.source, test.context); effect != nil {
				t.Fatalf("source %q was recognized as a dynamic source-spell cost reduction", test.source)
			}
		})
	}
}

// TestParseSourceSpellCostReductionHistoric covers the "historic" count
// qualifier (artifact, legendary, or Saga) on a graveyard count, including The
// Capitoline Triad's shape where the reduction carries an ability-word prefix and
// a reminder-text sentence.
func TestParseSourceSpellCostReductionHistoric(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		source  string
		context Context
	}{
		{
			name:    "plain historic graveyard count",
			source:  "This spell costs {1} less to cast for each historic card in your graveyard.",
			context: Context{InstantOrSorcery: true},
		},
		{
			name:    "ability word and reminder text",
			source:  "Those Who Came Before — This spell costs {1} less to cast for each historic card in your graveyard. (Artifacts, legendaries, and Sagas are historic.)",
			context: Context{InstantOrSorcery: true},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			effect := sourceSpellReductionEffect(t, test.source, test.context)
			if effect == nil {
				t.Fatalf("source %q did not yield a source-spell cost reduction", test.source)
			}
			if effect.Amount.Selection == nil {
				t.Fatalf("count selection missing for %q", test.source)
			}
			if !effect.Amount.Selection.Historic {
				t.Fatalf("historic flag not set for %q", test.source)
			}
		})
	}
}

// sourceSpellReductionConditionalEffect returns the first effect across an
// ability's sentences that is marked as a conditional source-spell cost
// reduction ("This spell costs {N} less to cast if <condition>.").
func sourceSpellReductionConditionalEffect(t *testing.T, source string, context Context) *EffectSyntax {
	t.Helper()
	document, _ := Parse(source, context)
	for ai := range document.Abilities {
		ability := &document.Abilities[ai]
		for si := range ability.Sentences {
			sentence := &ability.Sentences[si]
			for ei := range sentence.Effects {
				if sentence.Effects[ei].SourceSpellCostReductionConditional {
					return &sentence.Effects[ei]
				}
			}
		}
	}
	return nil
}

func TestParseSourceSpellCostReductionConditionalExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		source  string
		context Context
		amount  int
	}{
		{
			name:    "control a wizard",
			source:  "This spell costs {2} less to cast if you control a Wizard.",
			context: Context{InstantOrSorcery: true},
			amount:  2,
		},
		{
			name:    "control a giant",
			source:  "This spell costs {3} less to cast if you control a Giant.",
			context: Context{InstantOrSorcery: true},
			amount:  3,
		},
		{
			name:    "self name subject",
			source:  "Squash costs {3} less to cast if you control a Giant.",
			context: Context{InstantOrSorcery: true, CardName: "Squash"},
			amount:  3,
		},
		{
			name:    "opponent controls a permanent",
			source:  "This spell costs {1} less to cast if an opponent controls a green permanent.",
			context: Context{InstantOrSorcery: true},
			amount:  1,
		},
		{
			name:    "targets tapped creature",
			source:  "This spell costs {3} less to cast if it targets a tapped creature.",
			context: Context{InstantOrSorcery: true},
			amount:  3,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			effect := sourceSpellReductionConditionalEffect(t, test.source, test.context)
			if effect == nil {
				t.Fatalf("source %q did not yield a conditional source-spell cost reduction", test.source)
			}
			if effect.SourceSpellCostReductionAmount != test.amount {
				t.Fatalf("reduction amount = %d, want %d", effect.SourceSpellCostReductionAmount, test.amount)
			}
			if test.name == "targets tapped creature" && !effect.SourceSpellCostReductionTargetsTappedCreature {
				t.Fatal("tapped-creature target condition was not typed")
			}
		})
	}
}

func TestParseSourceSpellCostReductionConditionalFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		source  string
		context Context
	}{
		{
			name:    "increase wording",
			source:  "This spell costs {2} more to cast if you control a Wizard.",
			context: Context{InstantOrSorcery: true},
		},
		{
			name:    "no condition unconditional flat",
			source:  "This spell costs {2} less to cast.",
			context: Context{InstantOrSorcery: true},
		},
		{
			name:    "for each count form",
			source:  "This spell costs {1} less to cast for each creature you control.",
			context: Context{InstantOrSorcery: true},
		},
		{
			name:    "extra resolving clause",
			source:  "This spell costs {2} less to cast if you control a Wizard. Draw a card.",
			context: Context{InstantOrSorcery: true},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if effect := sourceSpellReductionConditionalEffect(t, test.source, test.context); effect != nil {
				t.Fatalf("source %q was recognized as a conditional source-spell cost reduction", test.source)
			}
		})
	}
}
