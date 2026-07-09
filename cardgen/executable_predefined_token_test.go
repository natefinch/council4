package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutablePredefinedMutavaultToken exercises the predefined
// named-token generator (issue #1560): Mutable Explorer's enter trigger creates
// a tapped Mutavault token whose identity is a card name rather than a card
// subtype. The token definition is fixed in lowering — a colorless land with the
// "{T}: Add {C}." mana ability and the "{1}: ... becomes a 2/2 creature with all
// creature types until end of turn. It's still a land." self-animation ability —
// since the create clause spells out only the name (the abilities live in the
// printed token's reminder text).
func TestGenerateExecutablePredefinedMutavaultToken(t *testing.T) {
	t.Parallel()
	power, toughness := "1", "1"
	card := &ScryfallCard{
		Name:      "Mutable Explorer",
		Layout:    "normal",
		ManaCost:  "{2}{G}",
		TypeLine:  "Creature — Shapeshifter",
		Power:     &power,
		Toughness: &toughness,
		OracleText: "Changeling (This card is every creature type.)\n" +
			"When this creature enters, create a tapped Mutavault token. " +
			"(It's a land with \"{T}: Add {C}\" and \"{1}: This token becomes a 2/2 creature with all creature types until end of turn. It's still a land.\")",
		Colors: []string{"G"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "m")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.CreateToken{",
		"Source:      game.TokenDef(mutableExplorerToken),",
		"EntryTapped: true,",
		`Name:  "Mutavault",`,
		"Types: []types.Card{types.Land},",
		"game.TapManaAbility(mana.C),",
		"Primitive: game.ApplyContinuous{",
		"Object: opt.Val(game.SourcePermanentReference()),",
		"AddEveryCreatureType: true,",
		"SetPower:     opt.Val(game.PT{Value: 2}),",
		"Duration: game.DurationUntilEndOfTurn,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutablePredefinedTarmogoyfToken exercises the predefined
// named-token generator for the Tarmogoyf token created by Disa the Restless: a
// green Lhurgoyf creature whose characteristic-defining ability sets its power to
// the number of card types among cards in all graveyards and its toughness to
// that number plus 1. The create clause carries only the name, so the token's
// full definition — its color, subtype, printed "*"/"*", and CDA
// DynamicPower/DynamicToughness — is fixed in lowering, reusing the same modeling
// the real Tarmogoyf card lowers to.
func TestGenerateExecutablePredefinedTarmogoyfToken(t *testing.T) {
	t.Parallel()
	power, toughness := "3", "3"
	card := &ScryfallCard{
		Name:       "Test Tarmogoyf Maker",
		Layout:     "normal",
		ManaCost:   "{1}{G}",
		TypeLine:   "Creature — Human Scout",
		Power:      &power,
		Toughness:  &toughness,
		OracleText: "When this creature enters, create a Tarmogoyf token.",
		Colors:     []string{"G"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.CreateToken{",
		"game.TokenDef(",
		`Name: "Tarmogoyf",`,
		"[]color.Color{color.Green}",
		"[]types.Sub{types.Lhurgoyf}",
		"game.PT{IsStar: true}",
		"game.DynamicValue{Kind: game.DynamicValueCardTypesAmongAllGraveyards}",
		"game.DynamicValue{Kind: game.DynamicValueCardTypesAmongAllGraveyards, Offset: 1}",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
