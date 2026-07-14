package parser

import "testing"

func graveyardEscapeGrantDeclaration(t *testing.T, source string) StaticDeclarationSyntax {
	t.Helper()
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	for _, ability := range document.Abilities {
		for _, declaration := range ability.StaticDeclarations {
			if declaration.Kind == StaticDeclarationGraveyardCardKeywordGrant {
				return declaration
			}
		}
	}
	t.Fatalf("Parse(%q) produced no graveyard keyword-grant declaration", source)
	return StaticDeclarationSyntax{}
}

// TestParseGraveyardEscapeGrantDeclaration proves the parser reads Underworld
// Breach's two-sentence escape grant into a nonland graveyard keyword-grant
// declaration carrying the typed computed escape cost.
func TestParseGraveyardEscapeGrantDeclaration(t *testing.T) {
	t.Parallel()
	declaration := graveyardEscapeGrantDeclaration(t,
		"Each nonland card in your graveyard has escape. The escape cost is equal to the card's mana cost plus exile three other cards from your graveyard.")
	if declaration.Subject.CardFilter != StaticDeclarationCardFilterNonland {
		t.Fatalf("filter = %v, want nonland", declaration.Subject.CardFilter)
	}
	if declaration.GraveyardEscapeCost == nil {
		t.Fatal("GraveyardEscapeCost = nil, want computed escape cost")
	}
	if !declaration.GraveyardEscapeCost.UseCardManaCost {
		t.Fatal("UseCardManaCost = false, want true")
	}
	if declaration.GraveyardEscapeCost.ExileOtherCount != 3 {
		t.Fatalf("ExileOtherCount = %d, want 3", declaration.GraveyardEscapeCost.ExileOtherCount)
	}
}

// TestParseGraveyardEscapeGrantFailsClosed proves the parser recognizes only the
// exact computed-escape-cost shape: any deviation yields no escape-cost
// declaration so the card fails closed downstream rather than lowering an
// approximate cost.
func TestParseGraveyardEscapeGrantFailsClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		// Exiles from the battlefield rather than the graveyard.
		"Each nonland card in your graveyard has escape. The escape cost is equal to the card's mana cost plus exile three other cards from the battlefield.",
		// Omits the "other" exclusion of the escaping card.
		"Each nonland card in your graveyard has escape. The escape cost is equal to the card's mana cost plus exile three cards from your graveyard.",
		// Cost clause pays life instead of exiling cards.
		"Each nonland card in your graveyard has escape. The escape cost is equal to the card's mana cost plus pay 3 life.",
	} {
		document, diagnostics := Parse(source, Context{})
		if len(diagnostics) != 0 {
			continue
		}
		for _, ability := range document.Abilities {
			for _, declaration := range ability.StaticDeclarations {
				if declaration.Kind == StaticDeclarationGraveyardCardKeywordGrant && declaration.GraveyardEscapeCost != nil {
					t.Fatalf("Parse(%q) produced a computed escape cost for an unsupported shape", source)
				}
			}
		}
	}
}
