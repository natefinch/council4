package parser

import "testing"

// TestParseChooseABackgroundRecognized confirms the parser recognizes the
// "Choose a Background" keyword ability, including its second-commander reminder,
// as a represented-but-not-simulated choose-a-background clause and clears the
// paragraph's competing effect, declaration, and condition semantics so
// downstream stages consume only the choose-a-background identity.
func TestParseChooseABackgroundRecognized(t *testing.T) {
	t.Parallel()
	source := "Choose a Background (You can have a Background as a second commander.)"
	document, diagnostics := Parse(source, Context{CardName: "Jaheira, Friend of the Forest"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %d, want 1", len(document.Abilities))
	}
	ability := document.Abilities[0]
	if ability.ChooseABackground == nil {
		t.Fatal("ChooseABackground clause = nil, want recognized choose-a-background ability")
	}
	if ability.Sentences != nil || ability.StaticDeclarations != nil ||
		ability.SemanticKeywords != nil || ability.ConditionClauses != nil ||
		ability.EventHistoryConditions != nil {
		t.Fatalf("competing semantics not cleared: %#v", ability)
	}
}
