package parser

import "testing"

// TestParseMultiTokenSharedVariableX proves Farmer Cotton's "create X 1/1 white
// Halfling creature tokens and X Food tokens." parses into one exact EffectCreate
// standing for the first spec (a 1/1 Halfling creature token) plus one
// AdditionalTokens entry for the predefined Food token, both carrying the shared
// variable X count. This is the reusable multi-token-with-shared-count parse: a
// synthesized creature spec and a predefined artifact spec joined under one X.
func TestParseMultiTokenSharedVariableX(t *testing.T) {
	t.Parallel()
	source := `When this creature enters, create X 1/1 white Halfling creature tokens and X Food tokens. (They're artifacts with "{2}, {T}, Sacrifice this token: You gain 3 life.")`
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse diagnostics = %#v", diagnostics)
	}
	effect := &document.Abilities[0].Sentences[0].Effects[0]
	if effect.Kind != EffectCreate || !effect.Exact {
		t.Fatalf("kind=%v exact=%v, want EffectCreate exact", effect.Kind, effect.Exact)
	}
	if len(effect.AdditionalTokens) != 1 {
		t.Fatalf("AdditionalTokens = %d, want 1", len(effect.AdditionalTokens))
	}
	if !effect.Amount.VariableX {
		t.Errorf("first spec amount = %+v, want VariableX", effect.Amount)
	}
	if !effect.TokenPTKnown ||
		len(effect.Selection.SubtypesAny) != 1 || effect.Selection.SubtypesAny[0] != "Halfling" {
		t.Errorf("first spec = %+v, want 1/1 Halfling creature token", effect.Selection)
	}
	food := effect.AdditionalTokens[0]
	if !food.Amount.VariableX {
		t.Errorf("Food spec amount = %+v, want VariableX", food.Amount)
	}
	if food.TokenPTKnown {
		t.Error("Food spec should carry no printed power/toughness, got PTKnown")
	}
	if len(food.Selection.SubtypesAny) != 1 || food.Selection.SubtypesAny[0] != "Food" {
		t.Errorf("Food spec subtypes = %v, want [Food]", food.Selection.SubtypesAny)
	}
}

// TestParseMultiTokenSharedFixedPredefined proves two predefined artifact tokens
// created together under a single fixed count (Madame Vastra's "create a Clue
// token and a Food token.") parse into one exact EffectCreate plus one
// AdditionalTokens entry, each a distinct predefined subtype with count 1.
func TestParseMultiTokenSharedFixedPredefined(t *testing.T) {
	t.Parallel()
	source := "Create a Clue token and a Food token."
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse diagnostics = %#v", diagnostics)
	}
	effect := &document.Abilities[0].Sentences[0].Effects[0]
	if effect.Kind != EffectCreate || !effect.Exact {
		t.Fatalf("kind=%v exact=%v, want EffectCreate exact", effect.Kind, effect.Exact)
	}
	if len(effect.AdditionalTokens) != 1 {
		t.Fatalf("AdditionalTokens = %d, want 1", len(effect.AdditionalTokens))
	}
	if effect.Amount.VariableX || !effect.Amount.Known || effect.Amount.Value != 1 {
		t.Errorf("first spec amount = %+v, want fixed 1", effect.Amount)
	}
	if len(effect.Selection.SubtypesAny) != 1 || effect.Selection.SubtypesAny[0] != "Clue" {
		t.Errorf("first spec subtypes = %v, want [Clue]", effect.Selection.SubtypesAny)
	}
	food := effect.AdditionalTokens[0]
	if food.Amount.VariableX || !food.Amount.Known || food.Amount.Value != 1 {
		t.Errorf("Food spec amount = %+v, want fixed 1", food.Amount)
	}
	if len(food.Selection.SubtypesAny) != 1 || food.Selection.SubtypesAny[0] != "Food" {
		t.Errorf("Food spec subtypes = %v, want [Food]", food.Selection.SubtypesAny)
	}
}

// TestParseMultiTokenMixedCountFailsClosed proves specs that do not share one
// representable count ("a 1/1 ... and X ... tokens") do not fold into the
// multi-token path: no AdditionalTokens are attached, so the clause cannot be
// mistaken for a shared-count multi-token create.
func TestParseMultiTokenMixedCountFailsClosed(t *testing.T) {
	t.Parallel()
	source := "Create a 1/1 green Saproling creature token and X Food tokens."
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse diagnostics = %#v", diagnostics)
	}
	effect := &document.Abilities[0].Sentences[0].Effects[0]
	if len(effect.AdditionalTokens) != 0 {
		t.Fatalf("AdditionalTokens = %d, want 0 (mixed counts must not fold)", len(effect.AdditionalTokens))
	}
}

func TestParseMultiTokenGrantedStaticAbilityBindsToMatchingSpec(t *testing.T) {
	t.Parallel()
	source := `Create a Treasure token and a 1/1 colorless Pilot creature token with "This creature crews Vehicles as though its power were 2 greater."`
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse diagnostics = %#v", diagnostics)
	}
	effect := &document.Abilities[0].Sentences[0].Effects[0]
	if len(effect.AdditionalTokens) != 1 {
		t.Fatalf("additional tokens = %d, exact=%v", len(effect.AdditionalTokens), effect.Exact)
	}
	if !effect.Exact {
		t.Fatalf("effect exact=false; first grant=%v pilot grant=%v root=%d..%d pilot=%d..%d quote=%d..%d",
			effect.TokenGrantedAbility != nil,
			effect.AdditionalTokens[0].TokenGrantedAbility != nil,
			effect.ClauseSpan.Start.Offset, effect.ClauseSpan.End.Offset,
			effect.AdditionalTokens[0].ClauseSpan.Start.Offset, effect.AdditionalTokens[0].ClauseSpan.End.Offset,
			document.Abilities[0].Quoted[0].Span.Start.Offset, document.Abilities[0].Quoted[0].Span.End.Offset)
	}
	if effect.TokenGrantedAbility != nil {
		t.Fatal("Treasure unexpectedly received Pilot ability")
	}
	pilot := &effect.AdditionalTokens[0]
	if pilot.TokenGrantedAbility == nil {
		t.Fatalf("Pilot spec = %#v, want granted static ability", pilot)
	}
	granted := pilot.TokenGrantedAbility.document.Abilities[0]
	if len(granted.StaticDeclarations) != 1 ||
		granted.StaticDeclarations[0].Kind != StaticDeclarationCrewPowerContribution ||
		granted.StaticDeclarations[0].CrewPowerBonus != 2 {
		t.Fatalf("granted static declarations = %#v", granted.StaticDeclarations)
	}
}
