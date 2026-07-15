package parser

import "testing"

// TestParsePluralVariableXTokenCount proves a plural variable-X token creation
// keeps its explicit token count on TokenCount while the "where X is <dynamic>"
// size clause rides on Amount, and the clause is recognized as exact so it
// lowers rather than falling back to the unsupported path.
func TestParsePluralVariableXTokenCount(t *testing.T) {
	t.Parallel()
	source := "Create three tapped X/X green Treefolk creature tokens, where X is the amount of life you gained this turn."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	effect := document.Abilities[0].Sentences[0].Effects[0]
	if !effect.Exact || effect.Kind != EffectCreate {
		t.Fatalf("effect = %#v, want an exact create", effect)
	}
	if !effect.TokenPTVariableX {
		t.Fatalf("effect = %#v, want variable X/X size", effect)
	}
	if !effect.TokenCount.Known || effect.TokenCount.Value != 3 {
		t.Fatalf("token count = %#v, want fixed 3", effect.TokenCount)
	}
	if effect.Amount.DynamicForm != EffectDynamicAmountFormWhereX ||
		effect.Amount.DynamicKind != EffectDynamicAmountLifeGainedThisTurn {
		t.Fatalf("size amount = %#v, want where-X life gained this turn", effect.Amount)
	}
	if !effect.Selection.Tapped {
		t.Fatalf("selection = %#v, want tapped entry", effect.Selection)
	}
}

// TestParseSingularVariableXTokenLeavesCountUnset proves the singular "an X/X
// ... where X is <dynamic>" form is byte-identical to before: TokenCount stays
// the zero amount so lowering defaults to a single token.
func TestParseSingularVariableXTokenLeavesCountUnset(t *testing.T) {
	t.Parallel()
	source := "Create an X/X green Ooze creature token, where X is the greatest power among creatures you control."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	effect := document.Abilities[0].Sentences[0].Effects[0]
	if !effect.Exact || !effect.TokenPTVariableX {
		t.Fatalf("effect = %#v, want an exact variable X/X create", effect)
	}
	if effect.TokenCount.Known || effect.TokenCount.Value != 0 {
		t.Fatalf("token count = %#v, want the zero amount for the singular form", effect.TokenCount)
	}
}

// TestParseCounterOnEachOfThem proves "Put a <kind> counter on each of them."
// is an exact plural-recipient counter placement, so the created-token counter
// sequence can place a counter on every member of the created group.
func TestParseCounterOnEachOfThem(t *testing.T) {
	t.Parallel()
	source := "Create three tapped X/X green Treefolk creature tokens, where X is the amount of life you gained this turn. Put a reach counter on each of them."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	counter := document.Abilities[0].Sentences[1].Effects[0]
	if !counter.Exact || counter.Kind != EffectPut {
		t.Fatalf("counter effect = %#v, want an exact put", counter)
	}
	if !counter.CounterKnown {
		t.Fatalf("counter effect = %#v, want a known counter kind", counter)
	}
	if len(counter.References) != 1 ||
		counter.References[0].Kind != ReferencePronoun ||
		counter.References[0].Pronoun != PronounThem {
		t.Fatalf("counter references = %#v, want the plural pronoun them", counter.References)
	}
}
