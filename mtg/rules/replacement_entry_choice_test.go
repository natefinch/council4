package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestEntryColorChoiceRecordedOnEntry verifies that "As this permanent enters,
// choose a color." prompts the controller as the permanent enters and records
// the chosen color on the permanent under EntryColorChoiceKey (CR 614.12).
func TestEntryColorChoiceRecordedOnEntry(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// The single-option prompt for a color lists W,U,B,R,G; select index 2 (B).
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{2}}}}

	def := &game.CardDef{CardFace: game.CardFace{
		Name:                 "Sol Grail",
		Types:                []types.Card{types.Artifact},
		ReplacementAbilities: []game.ReplacementAbility{game.EntryColorChoiceReplacement("As this artifact enters, choose a color.")},
		ManaAbilities:        []game.ManaAbility{game.TapChosenColorManaAbility("{T}: Add one mana of the chosen color.")},
	}}
	cardID := addCardToHand(g, game.Player1, def)
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		t.Fatal("card instance not found")
	}
	g.Players[game.Player1].Hand.Remove(cardID)

	permanent, ok := createCardPermanentWithChoices(engine, g, card, game.Player1, zone.Hand, agents, &TurnLog{})
	if !ok {
		t.Fatal("createCardPermanentWithChoices() = false, want true")
	}
	if permanent.Tapped {
		t.Fatal("standalone choose-a-color permanent must not enter tapped")
	}
	result, ok := permanent.EntryChoices[game.EntryColorChoiceKey]
	if !ok {
		t.Fatalf("entry color choice not recorded: %+v", permanent.EntryChoices)
	}
	if result.Color != mana.B {
		t.Fatalf("recorded entry color = %v, want %v", result.Color, mana.B)
	}

	// The mana ability must produce one mana of the color chosen on entry.
	want := action.ActivateAbility(permanent.ObjectID, 0, nil, 0)
	if !engine.applyAction(g, game.Player1, want) {
		t.Fatal("applyAction(chosen-color mana ability) = false, want true")
	}
	if !permanent.Tapped {
		t.Fatal("mana ability did not tap the permanent")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.B); got != 1 {
		t.Fatalf("chosen-color mana = %d, want 1 black", got)
	}
	if total := g.Players[game.Player1].ManaPool.Total(); total != 1 {
		t.Fatalf("mana pool total = %d, want exactly 1", total)
	}
}

// TestEntersTappedColorChoiceTapsAndRecords verifies the combined "This land
// enters tapped. As it enters, choose a color." both taps the permanent and
// records the chosen color on entry.
func TestEntersTappedColorChoiceTapsAndRecords(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// Select index 4 (G) from the color prompt.
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{4}}}}

	def := &game.CardDef{CardFace: game.CardFace{
		Name:                 "Uncharted Haven",
		Types:                []types.Card{types.Land},
		ReplacementAbilities: []game.ReplacementAbility{game.EntersTappedColorChoiceReplacement("This land enters tapped. As it enters, choose a color.")},
		ManaAbilities:        []game.ManaAbility{game.TapChosenColorManaAbility("{T}: Add one mana of the chosen color.")},
	}}
	cardID := addCardToHand(g, game.Player1, def)
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		t.Fatal("card instance not found")
	}
	g.Players[game.Player1].Hand.Remove(cardID)

	permanent, ok := createCardPermanentWithChoices(engine, g, card, game.Player1, zone.Hand, agents, &TurnLog{})
	if !ok {
		t.Fatal("createCardPermanentWithChoices() = false, want true")
	}
	if !permanent.Tapped {
		t.Fatal("combined enters-tapped color-choice permanent did not enter tapped")
	}
	result, ok := permanent.EntryChoices[game.EntryColorChoiceKey]
	if !ok {
		t.Fatalf("entry color choice not recorded: %+v", permanent.EntryChoices)
	}
	if result.Color != mana.G {
		t.Fatalf("recorded entry color = %v, want %v", result.Color, mana.G)
	}
}
